import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";
import { PDFDocument } from "pdf-lib";
 
/**
 * Genera y descarga un Expediente PDF completo para un documento radicado.
 * 
 * @param {Object} radicado - Información del radicado
 * @param {Array} flujo - Lista de tareas del flujo de aprobación
 * @param {Array} trazabilidad - Historial de eventos
 * @param {Array} anexosUrls - Lista de URLs de los anexos PDF para fusionar
 */
export async function generarExpedientePDF(radicado, flujo, trazabilidad, anexosUrls) {
  // 1. Generar Portada (Solo Normas de Reparto)
  const docPortada = new jsPDF();
  docPortada.setFontSize(18);
  docPortada.text(`Expediente: ${radicado.numero_radicado} - Normas de Reparto`, 14, 20);
  
  const formatCurrency = (val) => val ? new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', minimumFractionDigits: 0 }).format(val) : "$ 0";

  // ── NORMAS DE REPARTO ──
  if (radicado.normas_reparto && radicado.normas_reparto.length > 0) {
    
    // 1. Obtener el subtotal de la factura
    const subtotalFactura = parseFloat(radicado.documento_comercial?.subtotal) || 0;
    
    // 2. Preparar filas con el valor calculado
    let totalCalculado = 0;
    const normasData = radicado.normas_reparto.map(n => {
      const pct = parseFloat(n.porcentaje) || 0;
      const valorAplicado = subtotalFactura * (pct / 100);
      totalCalculado += valorAplicado;
      
      const proj = n.proyecto || n.norma_reparto?.proyecto || "";
      const rawDesc = n.descripcion || n.norma_reparto?.descripcion || "";
      const desc = rawDesc.length > 80 ? rawDesc.substring(0, 77) + "..." : rawDesc;
      
      return [
        n.norma_reparto?.nombre || n.nombre || "",
        n.norma_reparto?.departamento || n.departamento || "",
        proj,
        desc,
        `${pct.toFixed(2)}%`,
        formatCurrency(valorAplicado)
      ];
    });

    // 3. Agregar fila de totales
    normasData.push([
      "", "", "",
      { content: "Total Distribuido", styles: { fontStyle: "bold", halign: "right" } },
      "",
      { content: formatCurrency(totalCalculado), styles: { fontStyle: "bold", halign: "right" } }
    ]);

    autoTable(docPortada, {
      startY: 30,
      head: [["Norma", "Área", "Proyecto", "Descripción", "%", "Valor Aplicado"]],
      body: normasData,
      theme: 'grid',
      headStyles: { fillColor: [200, 160, 30] },
      styles: { fontSize: 8 },
      columnStyles: { 
        3: { cellWidth: 55 },
        4: { halign: 'right', fontStyle: 'bold' },
        5: { halign: 'right', fontStyle: 'bold' }
      }
    });
  } else {
    docPortada.setFontSize(12);
    docPortada.text("No se han asignado normas de reparto.", 14, 30);
  }

  const portadaBytes = docPortada.output('arraybuffer');

  // 2. Inicializar pdf-lib
  const mergedPdf = await PDFDocument.create();
  
  // Añadir portada (ahora incluye el QR)
  const portadaDoc = await PDFDocument.load(portadaBytes);
  const portadaPages = await mergedPdf.copyPages(portadaDoc, portadaDoc.getPageIndices());
  portadaPages.forEach((page) => mergedPdf.addPage(page));

  // 3. Añadir anexos (solo PDFs)
  for (const url of anexosUrls) {
    try {
      const fetchRes = await fetch(url, {
        cache: 'no-store',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
        }
      });
      if (fetchRes.ok) {
        const anexoBytes = await fetchRes.arrayBuffer();
        const anexoDoc = await PDFDocument.load(anexoBytes);
        const anexoPages = await mergedPdf.copyPages(anexoDoc, anexoDoc.getPageIndices());
        anexoPages.forEach((page) => mergedPdf.addPage(page));
      }
    } catch (err) {
      console.error("Error cargando anexo para expediente:", url, err);
    }
  }

  // 5. Guardar y forzar descarga
  const finalBytes = await mergedPdf.save();
  const blob = new Blob([finalBytes], { type: 'application/pdf' });
  const downloadUrl = URL.createObjectURL(blob);
  
  const a = document.createElement('a');
  a.href = downloadUrl;
  a.download = `Expediente_${radicado.numero_radicado}.pdf`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  setTimeout(() => URL.revokeObjectURL(downloadUrl), 60000);
}