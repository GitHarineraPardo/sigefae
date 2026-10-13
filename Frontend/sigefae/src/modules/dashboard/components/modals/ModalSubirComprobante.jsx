import { useState } from 'react';
import { API } from '../../constants/api';

export default function ModalSubirComprobante({
  isOpen, 
  onClose, 
  radicadoId, 
  obtenerToken,
  onUploadSuccess
}) {
  const [archivos, setArchivos] = useState([]);
  const [isUploading, setIsUploading] = useState(false);

  if (!isOpen) return null;

  const handleFileChange = (e) => {
    if (e.target.files) {
      setArchivos(prev => [...prev, ...Array.from(e.target.files)]);
    }
  };

  const removeFile = (indexToRemove) => {
    setArchivos(prev => prev.filter((_, idx) => idx !== indexToRemove));
  };

  const handleUpload = async () => {
    if (archivos.length === 0) {
      alert("Por favor, selecciona al menos un archivo.");
      return;
    }

    setIsUploading(true);
    try {
      const token = obtenerToken();
      
      // 1. Subir cada archivo (reutilizando el endpoint de anexos)
      for (const file of archivos) {
        const formData = new FormData();
        formData.append("file", file);
        const res = await fetch(`${API}/documentoradicado/${radicadoId}/anexos`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
          body: formData,
        });
        if (!res.ok) {
          const err = await res.json();
          throw new Error(err.error || "Error subiendo archivo");
        }
      }

      // 2. Marcar ComprobantesSubidos = true en el backend
      const resLock = await fetch(`${API}/documentoradicado/${radicadoId}/comprobantes_subidos`, {
        method: "PATCH",
        headers: { 
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}` 
        },
        body: JSON.stringify({ comprobantes_subidos: true })
      });

      if (!resLock.ok) {
        throw new Error("Error bloqueando la subida de comprobantes");
      }

      alert("Comprobantes subidos exitosamente.");
      onUploadSuccess();
      onClose();
    } catch (err) {
      alert("Error: " + err.message);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content" style={{ maxWidth: '500px' }}>
        <div className="modal-header">
          <h2><i className="fa-solid fa-file-invoice-dollar"></i> Subir Comprobante(s) de Pago</h2>
          <button className="btn-close" onClick={onClose} disabled={isUploading}>
            <i className="fa-solid fa-xmark"></i>
          </button>
        </div>
        
        <div className="modal-body" style={{ display: 'flex', flexDirection: 'column', gap: '15px' }}>
          <p style={{ fontSize: '0.9em', color: '#64748b' }}>
            Selecciona uno o más archivos. Una vez subidos, esta opción se bloqueará y solo el Superadministrador podrá revertirlo.
          </p>

          <div style={{ border: '2px dashed #cbd5e1', padding: '20px', borderRadius: '8px', textAlign: 'center', backgroundColor: '#f8fafc' }}>
            <input 
              type="file" 
              multiple 
              onChange={handleFileChange}
              style={{ display: 'none' }}
              id="comprobante-upload-input"
              disabled={isUploading}
            />
            <label 
              htmlFor="comprobante-upload-input" 
              className="doc-btn doc-btn-secondary" 
              style={{ cursor: isUploading ? 'not-allowed' : 'pointer' }}
            >
              <i className="fa-solid fa-folder-open"></i> Seleccionar Archivos
            </label>
            
            {archivos.length > 0 && (
              <div style={{ marginTop: '15px', textAlign: 'left' }}>
                <p style={{ fontSize: '0.85em', color: '#0f172a', fontWeight: '600', marginBottom: '8px' }}>
                  {archivos.length} archivo(s) seleccionado(s):
                </p>
                <ul style={{ listStyle: 'none', padding: 0, margin: 0, display: 'flex', flexDirection: 'column', gap: '6px' }}>
                  {archivos.map((file, idx) => (
                    <li key={idx} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#e2e8f0', padding: '6px 10px', borderRadius: '4px', fontSize: '0.85em', color: '#334155' }}>
                      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '80%' }}>
                        <i className="fa-solid fa-file" style={{ marginRight: '6px', color: '#64748b' }}></i>
                        {file.name}
                      </span>
                      <button 
                        onClick={() => removeFile(idx)} 
                        disabled={isUploading}
                        style={{ background: 'none', border: 'none', color: '#ef4444', cursor: isUploading ? 'not-allowed' : 'pointer', padding: '2px 4px' }}
                        title="Quitar archivo"
                      >
                        <i className="fa-solid fa-xmark"></i>
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </div>

        <div className="modal-footer" style={{ marginTop: '20px', display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
          <button className="doc-btn doc-btn-secondary" onClick={onClose} disabled={isUploading}>
            Cancelar
          </button>
          <button className="doc-btn doc-btn-primary" onClick={handleUpload} disabled={isUploading || archivos.length === 0}>
            {isUploading ? <><i className="fa-solid fa-spinner fa-spin"></i> Subiendo...</> : <><i className="fa-solid fa-cloud-arrow-up"></i> Confirmar Subida</>}
          </button>
        </div>
      </div>
    </div>
  );
}
