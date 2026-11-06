package constant

const (
	FileTypePDF       = "application/pdf"
	FileTypeImagePng  = "image/png"
	FileTypeImageJpeg = "image/jpeg"
	FileTypeImageJpg  = "image/jpg"
	FileTypeVideoWebm = "video/webm"
)

var ValidFileUploadTypes = []string{FileTypePDF, FileTypeImageJpeg, FileTypeImagePng, FileTypeImageJpg, FileTypeVideoWebm}

func GetFileExtFromMime(mime string) string {
	switch mime {
	case FileTypePDF:
		return "pdf"
	case FileTypeImageJpeg:
		return "jpeg"

	case FileTypeImagePng:
		return "png"
	case FileTypeImageJpg:
		return "jpg"
	case FileTypeVideoWebm:
		return "webm"
	}
	return ""
}
