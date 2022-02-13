package image

var (
	SGBackend = new(SG)
)

type SG struct {
	picBed
}

// func (s SG) linkExtractor(_ string) string {
// 	panic("linkExtractor is temporally unavailable on SoGou API")
// }

func (s SG) linkBuilder(link string) string {
	return link
}

func (s SG) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "http://pic.sogou.com/pic/upload_pic.jsp", "pic_path", defaultReqMod)
	if err != nil {
		return "", err
	}
	return s.linkBuilder(string(body)), nil
}
