import { API } from "../dashboard/constants/api.js";

export async function login(email, password, idExterno) {

    const response = await fetch(`${API}/auth/login`, {

        method: "POST",

        headers: {
            "Content-Type": "application/json"
        },

        body: JSON.stringify({
            ...(email ? { email } : {}),
            ...(idExterno ? { id_externo: idExterno } : {}),
            password
        })

    });

    const data = await response.json();

    if (!response.ok) {
        throw new Error(data.error || "Error al iniciar sesión");
    }

    return data;
}