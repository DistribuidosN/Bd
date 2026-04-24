package utils

// Función auxiliar para mapear el MIME type a la extensión
func GetExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	default:
		return ".bin" // Fallback seguro por si envían basura que no es imagen
	}
}