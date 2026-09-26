import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { guardarToken, obtenerToken } from "../auth/token";
import { API } from "../dashboard/constants/api";
import "./login.css";

function Login() {
    const navigate = useNavigate();

    useEffect(() => {
        const hydrateUser = async (sessionToken) => {
            localStorage.removeItem("user_cargo_id");
            try {
                const response = await fetch(`${API}/me`, { headers: { Authorization: `Bearer ${sessionToken}` } });
                if (response.ok) {
                    const user = await response.json();
                    if (user.cargo_id) localStorage.setItem("user_cargo_id", user.cargo_id);
                    else localStorage.removeItem("user_cargo_id");
                    if (user.cargo) localStorage.setItem("user_cargo", user.cargo);
                }
            } catch (error) {
                console.error("No se pudo cargar el cargo del usuario:", error);
            }
            navigate("/dashboard");
        };

        const params = new URLSearchParams(window.location.search);
        const token = params.get("sso_token");
        const rol = params.get("sso_rol");
        const userId = params.get("sso_user_id");
        const nombre = params.get("sso_nombre");
        const cargo = params.get("sso_cargo");

        // 1. Si vienen parámetros SSO en la URL, iniciar sesión automáticamente
        if (token && rol && userId) {
            guardarToken(token);
            localStorage.setItem("rol", rol);
            localStorage.setItem("user_id", userId);
            localStorage.removeItem("user_cargo_id");
            if (nombre) localStorage.setItem("user_name", nombre);
            if (cargo) localStorage.setItem("user_cargo", cargo);
            
            // Limpiar la URL y cargar el perfil antes de abrir el panel.
            window.history.replaceState({}, document.title, window.location.pathname);
            void hydrateUser(token);
            return;
        }

        // 2. Si no hay parámetros en la URL, verificar si ya hay una sesión activa
        const currentToken = obtenerToken();
        if (currentToken) {
            void hydrateUser(currentToken);
            return;
        }

        // 3. Si no hay parámetros ni sesión activa, devolver al panel principal
        window.location.href = "https://app.harinerapardo.co/index.php";

    }, [navigate]);

    return (
        <div className="login-container" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', backgroundColor: '#f0f2f5' }}>
            <div style={{ textAlign: 'center' }}>
                <div className="spinner" style={{ width: '40px', height: '40px', border: '4px solid rgba(0,0,0,0.1)', borderLeftColor: '#0056b3', borderRadius: '50%', animation: 'spin 1s linear infinite', margin: '0 auto 20px' }}></div>
                <h3 style={{ color: '#333', fontFamily: 'sans-serif' }}>Autenticando...</h3>
                <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
            </div>
        </div>
    );
}

export default Login;