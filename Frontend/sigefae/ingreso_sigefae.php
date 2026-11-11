<?php
if (session_status() === PHP_SESSION_NONE) {
    session_start();
}

if (empty($_SESSION['usuario_id'])) {
    die('No estás logueado en el panel o no se encontró tu ID de usuario.');
}

// Debe coincidir con usuario.id_externo en SIGEFAE.
$idExterno = trim((string) $_SESSION['usuario_id']);

$payload = json_encode([
    'id_externo' => $idExterno,
    'secret' => 'SIGEFAE_INTERNAL_SSO_SECRET_2026',
]);

if ($payload === false) {
    die('No se pudo preparar la solicitud de inicio de sesión.');
}

$ch = curl_init('http://127.0.0.1:8080/api/auth/sso');

curl_setopt_array($ch, [
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
    CURLOPT_POST => true,
    CURLOPT_POSTFIELDS => $payload,
    CURLOPT_CONNECTTIMEOUT => 5,
    CURLOPT_TIMEOUT => 15,
]);

$response = curl_exec($ch);
$curlError = curl_error($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

if ($response === false) {
    die('Error de conexión con el backend en Go: ' . htmlspecialchars($curlError, ENT_QUOTES, 'UTF-8'));
}

if ($httpCode !== 200) {
    die(
        'SIGEFAE no encontró una cuenta vinculada al ID externo. ' .
        'Verifica que usuario.id_externo coincida con el ID de sesión del panel. Respuesta: ' .
        htmlspecialchars($response, ENT_QUOTES, 'UTF-8')
    );
}

$data = json_decode($response, true);

if (!is_array($data) || empty($data['token']) || empty($data['usuario']['id'])) {
    die('La respuesta de SIGEFAE no contiene un token o usuario válido.');
}

$query = http_build_query([
    'sso_token' => $data['token'],
    'sso_rol' => $data['usuario']['rol'] ?? '',
    'sso_user_id' => $data['usuario']['id'],
    'sso_nombre' => $data['usuario']['nombre'] ?? '',
], '', '&', PHP_QUERY_RFC3986);

header('Location: Frontend/?' . $query);
exit;