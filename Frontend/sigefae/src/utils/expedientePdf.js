import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";
import { PDFDocument } from "pdf-lib";

const EVENTOS_CONFIG = [
  {
    codigo: "030",
    tipo: "ACKNOWLEDGED",
    inicial: "AC",
    nombre: "Acuse de Recibo (030)",
    color: [37, 99, 235], // #2563eb
    side: "left",
  },
  {
    codigo: "032",
    tipo: "RECEIVEDGOODS",
    inicial: "RB",
    nombre: "Recibo Bienes / Servicios (032)",
    color: [217, 119, 6], // #d97706
    side: "right",
  },
  {
    codigo: "034",
    tipo: "ACCEPTED",
    inicial: "AE",
    nombre: "Aceptación Expresa (034)",
    color: [22, 163, 74], // #16a34a
    side: "left",
  },
  {
    codigo: "031",
    tipo: "REJECTED",
    inicial: "RC",
    nombre: "Reclamo / Rechazo (031)",
    color: [220, 38, 38], // #dc2626
    side: "right",
  },
];

const formatCurrency = (val) =>
  val != null
    ? new Intl.NumberFormat("es-CO", {
        style: "currency",
        currency: "COP",
        minimumFractionDigits: 0,
        maximumFractionDigits: 2,
      }).format(val)
    : "$ 0";

/**
 * Genera y descarga el Expediente PDF completo para un documento radicado.
 *
 * Estructura y orden del documento:
 * 1. Factura y anexos PDF
 * 2. Activos Fijos y Normas de Reparto
 * 3. (En esa misma hoja si caben): Eventos Electrónicos DIAN en forma de arbolito/timeline.
 */
export async function generarExpedientePDF(
  radicadoOrOptions,
  flujoArg,
  trazabilidadArg,
  anexosUrlsArg,
  activosFijosArg,
  eventosDianArg
) {
  let radicado, anexosUrls, activosFijos, eventosDian;

  if (
    radicadoOrOptions &&
    typeof radicadoOrOptions === "object" &&
    ("radicado" in radicadoOrOptions || "anexosUrls" in radicadoOrOptions)
  ) {
    radicado = radicadoOrOptions.radicado || {};
    anexosUrls = radicadoOrOptions.anexosUrls || [];
    activosFijos = radicadoOrOptions.activosFijos || radicado.activos_fijos || [];
    eventosDian = radicadoOrOptions.eventosDian || [];
  } else {
    radicado = radicadoOrOptions || {};
    anexosUrls = anexosUrlsArg || [];
    activosFijos = activosFijosArg || radicado.activos_fijos || [];
    eventosDian = eventosDianArg || [];
  }

  // ─────────────────────────────────────────────────────────────
  // 1. INICIALIZAR PDF FINAL CON PDF-LIB
  // ─────────────────────────────────────────────────────────────
  const mergedPdf = await PDFDocument.create();

  // 1. Factura y anexos PDF PRIMERO
  for (const url of anexosUrls) {
    try {
      const fetchRes = await fetch(url, {
        cache: "no-store",
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token") || ""}`,
        },
      });
      if (fetchRes.ok) {
        const anexoBytes = await fetchRes.arrayBuffer();
        const anexoDoc = await PDFDocument.load(anexoBytes);
        const anexoPages = await mergedPdf.copyPages(
          anexoDoc,
          anexoDoc.getPageIndices()
        );
        anexoPages.forEach((page) => mergedPdf.addPage(page));
      }
    } catch (err) {
      console.error("Error cargando anexo para expediente:", url, err);
    }
  }

  // ─────────────────────────────────────────────────────────────
  // 2. GENERAR HOJA DE RESUMEN: ACTIVOS, NORMAS Y EVENTOS DIAN
  // ─────────────────────────────────────────────────────────────
  const doc = new jsPDF({ orientation: "portrait", unit: "mm", format: "a4" });

  const subtotalFactura =
    parseFloat(radicado.documento_comercial?.subtotal) || 0;
  const totalFactura = parseFloat(radicado.documento_comercial?.total) || 0;

  // ── Encabezado Principal ──
  doc.setFillColor(15, 23, 42); // slate-900
  doc.rect(14, 12, 182, 1.5, "F");

  doc.setFont("helvetica", "bold");
  doc.setFontSize(13);
  doc.setTextColor(15, 23, 42);
  doc.text(
    `EXPEDIENTE DE RADICACIÓN: ${radicado.numero_radicado || "S/N"}`,
    14,
    20
  );

  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(100, 116, 139);
  doc.text(`Sistema de Gestión de Facturas Electrónicas — SIGEFAE`, 14, 24.5);
  doc.text(
    `Fecha Emisión: ${new Date().toLocaleString("es-CO")}`,
    196,
    24.5,
    { align: "right" }
  );

  // ── Metadatos Generales ──
  doc.setFillColor(248, 250, 252);
  doc.setDrawColor(226, 232, 240);
  doc.setLineWidth(0.3);
  doc.roundedRect(14, 27, 182, 18, 1.5, 1.5, "FD");

  doc.setFontSize(7.5);
  doc.setFont("helvetica", "bold");
  doc.setTextColor(71, 85, 105);
  doc.text("Proveedor:", 17, 32);
  doc.text("Factura:", 17, 37);
  doc.text("F. Radicación:", 17, 42);

  doc.setFont("helvetica", "normal");
  doc.setTextColor(15, 23, 42);
  const provRazon =
    radicado.documento_comercial?.proveedor?.razon_social ||
    radicado.documento_comercial?.proveedor?.nombre ||
    "—";
  const provNit = radicado.documento_comercial?.proveedor?.numero_documento
    ? ` · NIT ${radicado.documento_comercial.proveedor.numero_documento}`
    : "";
  const provFull = `${provRazon}${provNit}`;
  doc.text(
    provFull.length > 50 ? provFull.substring(0, 48) + "..." : provFull,
    36,
    32
  );
  doc.text(radicado.documento_comercial?.numero_documento || "—", 36, 37);
  doc.text(
    radicado.fecha_radicacion
      ? new Date(radicado.fecha_radicacion).toLocaleDateString("es-CO")
      : "—",
    36,
    42
  );

  doc.setFont("helvetica", "bold");
  doc.setTextColor(71, 85, 105);
  doc.text("Subtotal:", 116, 32);
  doc.text("Total Factura:", 116, 37);
  doc.text("Ruta:", 116, 42);

  doc.setFont("helvetica", "normal");
  doc.setTextColor(15, 23, 42);
  doc.text(formatCurrency(subtotalFactura), 146, 32);
  doc.text(formatCurrency(totalFactura), 146, 37);
  const rutaStr = `${radicado.ruta?.nombre || "General"}${
    radicado.tipo_radicacion?.nombre ? " (" + radicado.tipo_radicacion.nombre + ")" : ""
  }`;
  doc.text(
    rutaStr.length > 32 ? rutaStr.substring(0, 30) + "..." : rutaStr,
    146,
    42
  );

  let currentY = 50;
  let totalActivosPct = 0;
  let totalActivosVal = 0;
  let totalNormasPct = 0;
  let totalNormasVal = 0;

  // ─────────────────────────────────────────────────────────────
  // 2. ACTIVOS FIJOS Y NORMAS DE REPARTO
  // ─────────────────────────────────────────────────────────────
  const activos = activosFijos || [];
  if (activos.length > 0) {
    doc.setFont("helvetica", "bold");
    doc.setFontSize(9.5);
    doc.setTextColor(15, 23, 42);
    doc.text(`1. Activos Fijos (${activos.length})`, 14, currentY);
    currentY += 2;

    const activosBody = activos.map((a) => {
      const pct = parseFloat(a.porcentaje) || 0;
      const val = subtotalFactura * (pct / 100);
      totalActivosPct += pct;
      totalActivosVal += val;
      const desc =
        a.descripcion && a.descripcion.length > 55
          ? a.descripcion.substring(0, 52) + "..."
          : a.descripcion || "—";
      return [
        a.nombre || "—",
        a.sucursal || "—",
        a.proyecto || "—",
        desc,
        `${pct.toFixed(2)}%`,
        formatCurrency(val),
      ];
    });

    activosBody.push([
      "",
      "",
      "",
      {
        content: "Total Activos Fijos",
        styles: { fontStyle: "bold", halign: "right" },
      },
      {
        content: `${totalActivosPct.toFixed(2)}%`,
        styles: { fontStyle: "bold", halign: "right" },
      },
      {
        content: formatCurrency(totalActivosVal),
        styles: { fontStyle: "bold", halign: "right" },
      },
    ]);

    autoTable(doc, {
      startY: currentY,
      head: [["Nombre", "Sucursal", "Proyecto", "Descripción", "%", "Valor Aplicado"]],
      body: activosBody,
      theme: "grid",
      headStyles: { fillColor: [44, 62, 80], fontSize: 7.5, fontStyle: "bold" },
      styles: { fontSize: 7, cellPadding: 1.8 },
      columnStyles: {
        0: { cellWidth: 32 },
        1: { cellWidth: 26 },
        2: { cellWidth: 26 },
        3: { cellWidth: 44 },
        4: { halign: "right", fontStyle: "bold", cellWidth: 18 },
        5: { halign: "right", fontStyle: "bold", cellWidth: 36 },
      },
      margin: { left: 14, right: 14 },
    });

    currentY = doc.lastAutoTable.finalY + 6;
  }

  const normas = radicado.normas_reparto || [];
  if (normas.length > 0) {
    doc.setFont("helvetica", "bold");
    doc.setFontSize(9.5);
    doc.setTextColor(15, 23, 42);
    const secNum = activos.length > 0 ? "2" : "1";
    doc.text(`${secNum}. Normas de Reparto (${normas.length})`, 14, currentY);
    currentY += 2;

    const normasBody = normas.map((n) => {
      const pct = parseFloat(n.porcentaje) || 0;
      const val = subtotalFactura * (pct / 100);
      totalNormasPct += pct;
      totalNormasVal += val;
      const proj = n.proyecto || n.norma_reparto?.proyecto || "—";
      const rawDesc = n.descripcion || n.norma_reparto?.descripcion || "—";
      const desc =
        rawDesc.length > 55 ? rawDesc.substring(0, 52) + "..." : rawDesc;
      return [
        n.norma_reparto?.codigo || n.codigo || "—",
        n.norma_reparto?.departamento || n.departamento || "—",
        proj,
        desc,
        `${pct.toFixed(2)}%`,
        formatCurrency(val),
      ];
    });

    normasBody.push([
      "",
      "",
      "",
      {
        content: "Total Normas de Reparto",
        styles: { fontStyle: "bold", halign: "right" },
      },
      {
        content: `${totalNormasPct.toFixed(2)}%`,
        styles: { fontStyle: "bold", halign: "right" },
      },
      {
        content: formatCurrency(totalNormasVal),
        styles: { fontStyle: "bold", halign: "right" },
      },
    ]);

    autoTable(doc, {
      startY: currentY,
      head: [["N.º Norma", "Área", "Proyecto", "Descripción", "%", "Valor Aplicado"]],
      body: normasBody,
      theme: "grid",
      headStyles: { fillColor: [200, 160, 30], fontSize: 7.5, fontStyle: "bold" },
      styles: { fontSize: 7, cellPadding: 1.8 },
      columnStyles: {
        0: { cellWidth: 26 },
        1: { cellWidth: 32 },
        2: { cellWidth: 26 },
        3: { cellWidth: 44 },
        4: { halign: "right", fontStyle: "bold", cellWidth: 18 },
        5: { halign: "right", fontStyle: "bold", cellWidth: 36 },
      },
      margin: { left: 14, right: 14 },
    });

    currentY = doc.lastAutoTable.finalY + 6;
  }

  if (activos.length === 0 && normas.length === 0) {
    doc.setFont("helvetica", "normal");
    doc.setFontSize(8);
    doc.setTextColor(100, 116, 139);
    doc.text(
      "No se asignaron normas de reparto ni activos fijos a este documento.",
      14,
      currentY
    );
    currentY += 8;
  } else if (activos.length > 0 && normas.length > 0) {
    const totalCombPct = totalActivosPct + totalNormasPct;
    const totalCombVal = totalActivosVal + totalNormasVal;
    doc.setFillColor(241, 245, 249);
    doc.roundedRect(14, currentY - 3, 182, 7, 1, 1, "F");
    doc.setFont("helvetica", "bold");
    doc.setFontSize(7.5);
    doc.setTextColor(15, 23, 42);
    doc.text(
      "DISTRIBUCIÓN TOTAL COMBINADA (Activos Fijos + Normas):",
      18,
      currentY + 1.8
    );
    doc.text(
      `${totalCombPct.toFixed(2)}%  —  ${formatCurrency(totalCombVal)}`,
      192,
      currentY + 1.8,
      { align: "right" }
    );
    currentY += 8;
  }

  // ─────────────────────────────────────────────────────────────
  // 3. EVENTOS ELECTRÓNICOS DIAN EN FORMA DE ARBOLITO
  // ─────────────────────────────────────────────────────────────
  const eventosMap = {};
  (eventosDian || []).forEach((e) => {
    if (e.codigo_dian) eventosMap[e.codigo_dian] = e;
    if (e.tipo_evento) eventosMap[e.tipo_evento.toUpperCase()] = e;
  });

  // Inferencia RADIAN
  const has030 = Boolean(eventosMap["030"] || eventosMap["ACKNOWLEDGED"]);
  const has032 = Boolean(eventosMap["032"] || eventosMap["RECEIVEDGOODS"]);
  const has034 = Boolean(eventosMap["034"] || eventosMap["ACCEPTED"]);
  const has031 = Boolean(eventosMap["031"] || eventosMap["REJECTED"]);

  if (!has030 && (has032 || has034 || has031)) {
    const ref =
      eventosMap["032"] || eventosMap["034"] || eventosMap["031"] || {};
    const synthetic = {
      codigo_dian: "030",
      tipo_evento: "ACKNOWLEDGED",
      fecha_envio: ref.fecha_envio || ref.created_at || new Date().toISOString(),
      usuario_nombre: "Buzón Inbound (Automático)",
      observacion:
        "Acuse de recibo generado automáticamente por el proveedor tecnológico al recibir la factura",
    };
    eventosMap["030"] = synthetic;
    eventosMap["ACKNOWLEDGED"] = synthetic;
  }

  if (!has032 && (has034 || has031)) {
    const ref = eventosMap["034"] || eventosMap["031"] || {};
    const synthetic = {
      codigo_dian: "032",
      tipo_evento: "RECEIVEDGOODS",
      fecha_envio: ref.fecha_envio || ref.created_at || new Date().toISOString(),
      usuario_nombre: "Buzón Inbound (Automático)",
      observacion:
        "Recibo de bienes o prestación de servicios registrado previamente ante la DIAN",
    };
    eventosMap["032"] = synthetic;
    eventosMap["RECEIVEDGOODS"] = synthetic;
  }

  // Verificar si cabe en esta misma hoja (requiere ~72mm de altura)
  const timelineHeight = 72;
  if (currentY + timelineHeight > 280) {
    doc.addPage();
    currentY = 20;
  }

  const secEventosNum =
    activos.length > 0 && normas.length > 0
      ? "3"
      : activos.length > 0 || normas.length > 0
      ? "2"
      : "1";

  doc.setFont("helvetica", "bold");
  doc.setFontSize(9.5);
  doc.setTextColor(15, 23, 42);
  doc.text(`${secEventosNum}. Eventos Electrónicos DIAN (RADIAN)`, 14, currentY);
  currentY += 8;

  const startTimelineY = currentY + 4;
  const rowGap = 18; // Separación entre nodos del arbolito
  const lastNodeY = startTimelineY + 3 * rowGap;

  // Eje central continuo (tronco del arbolito)
  doc.setDrawColor(203, 213, 225); // slate-300
  doc.setLineWidth(0.7);
  doc.line(105, startTimelineY, 105, lastNodeY);

  // Renderizar cada uno de los 4 eventos en el timeline alternado
  EVENTOS_CONFIG.forEach((cfg, idx) => {
    const nodeY = startTimelineY + idx * rowGap;
    const evento = eventosMap[cfg.codigo] || eventosMap[cfg.tipo];
    const isActivo = Boolean(evento);
    const isLeft = cfg.side === "left";

    const cardW = 74;
    const cardX = isLeft ? 105 - 12 - cardW : 105 + 12;
    const connX1 = isLeft ? cardX + cardW : 105 + 4.2;
    const connX2 = isLeft ? 105 - 4.2 : cardX;

    if (isActivo) {
      // Conector activo
      doc.setDrawColor(cfg.color[0], cfg.color[1], cfg.color[2]);
      doc.setLineWidth(0.6);
      doc.line(connX1, nodeY, connX2, nodeY);

      // Tarjeta de información del evento
      const obs = evento.observacion ? `"${evento.observacion}"` : "";
      const obsLines = obs ? doc.splitTextToSize(obs, cardW - 8) : [];
      const clampedObs = obsLines.slice(0, 2);
      const cardH = clampedObs.length > 0 ? 15.5 : 12;
      const cardY = nodeY - cardH / 2;

      // Caja de la tarjeta
      doc.setFillColor(248, 250, 252);
      doc.setDrawColor(226, 232, 240);
      doc.setLineWidth(0.3);
      doc.roundedRect(cardX, cardY, cardW, cardH, 1.2, 1.2, "FD");

      // Borde de acento de color en el lateral hacia el nodo
      doc.setFillColor(cfg.color[0], cfg.color[1], cfg.color[2]);
      if (isLeft) {
        doc.rect(cardX + cardW - 1.6, cardY, 1.6, cardH, "F");
      } else {
        doc.rect(cardX, cardY, 1.6, cardH, "F");
      }

      // Contenido de la tarjeta
      const textPadX = isLeft ? cardX + 3 : cardX + 4;
      doc.setFont("helvetica", "bold");
      doc.setFontSize(7.5);
      doc.setTextColor(15, 23, 42);
      doc.text(cfg.nombre, textPadX, cardY + 3.8);

      const fechaRaw = evento.fecha_envio || evento.created_at;
      const fechaStr = fechaRaw
        ? new Date(fechaRaw).toLocaleString("es-CO", {
            dateStyle: "short",
            timeStyle: "short",
          })
        : "Registrado";
      const userStr = evento.usuario_nombre ? ` · ${evento.usuario_nombre}` : "";
      doc.setFont("helvetica", "normal");
      doc.setFontSize(6.5);
      doc.setTextColor(100, 116, 139);
      doc.text(`${fechaStr}${userStr}`, textPadX, cardY + 7.5);

      if (clampedObs.length > 0) {
        doc.setFont("helvetica", "italic");
        doc.setFontSize(5.8);
        doc.setTextColor(71, 85, 105);
        doc.text(clampedObs, textPadX, cardY + 11);
      }

      // Bolita central activa (círculo de color + siglas blancas)
      doc.setFillColor(cfg.color[0], cfg.color[1], cfg.color[2]);
      doc.setDrawColor(255, 255, 255);
      doc.setLineWidth(0.5);
      doc.circle(105, nodeY, 4.2, "FD");

      doc.setFont("helvetica", "bold");
      doc.setFontSize(7);
      doc.setTextColor(255, 255, 255);
      doc.text(cfg.inicial, 105, nodeY + 1.2, { align: "center" });
    } else {
      // Conector pendiente
      doc.setDrawColor(226, 232, 240);
      doc.setLineWidth(0.3);
      doc.line(connX1, nodeY, connX2, nodeY);

      // Caja discreta pendiente
      const cardH = 7.5;
      const cardY = nodeY - cardH / 2;
      doc.setFillColor(255, 255, 255);
      doc.setDrawColor(226, 232, 240);
      doc.setLineWidth(0.2);
      doc.roundedRect(cardX, cardY, cardW, cardH, 1, 1, "FD");

      const textPadX = isLeft ? cardX + 3 : cardX + 4;
      doc.setFont("helvetica", "normal");
      doc.setFontSize(6.8);
      doc.setTextColor(148, 163, 184);
      doc.text(`${cfg.nombre} — Pendiente`, textPadX, nodeY + 1);

      // Bolita central pendiente
      doc.setFillColor(241, 245, 249);
      doc.setDrawColor(203, 213, 225);
      doc.setLineWidth(0.4);
      doc.circle(105, nodeY, 3.6, "FD");

      doc.setFont("helvetica", "bold");
      doc.setFontSize(6.2);
      doc.setTextColor(148, 163, 184);
      doc.text(cfg.inicial, 105, nodeY + 1.1, { align: "center" });
    }
  });

  // ─────────────────────────────────────────────────────────────
  // 4. INCORPORAR LA HOJA DE RESUMEN AL EXPEDIENTE FINAL
  // ─────────────────────────────────────────────────────────────
  const resumenBytes = doc.output("arraybuffer");
  const resumenPdfDoc = await PDFDocument.load(resumenBytes);
  const resumenPages = await mergedPdf.copyPages(
    resumenPdfDoc,
    resumenPdfDoc.getPageIndices()
  );
  resumenPages.forEach((page) => mergedPdf.addPage(page));

  // ─────────────────────────────────────────────────────────────
  // 5. GUARDAR Y FORZAR DESCARGA
  // ─────────────────────────────────────────────────────────────
  const finalBytes = await mergedPdf.save();
  const blob = new Blob([finalBytes], { type: "application/pdf" });
  const downloadUrl = URL.createObjectURL(blob);

  const a = document.createElement("a");
  a.href = downloadUrl;
  a.download = `Expediente_${radicado.numero_radicado || "radicado"}.pdf`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  setTimeout(() => URL.revokeObjectURL(downloadUrl), 60000);
}