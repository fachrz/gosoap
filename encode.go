package gosoap

import (
	"encoding/xml"
)

var start = xml.StartElement{
	Name: xml.Name{
		Space: "",
		Local: "soap:Envelope",
	},
	Attr: []xml.Attr{
		{Name: xml.Name{Space: "", Local: "xmlns:xsi"}, Value: "http://www.w3.org/2001/XMLSchema-instance"},
		{Name: xml.Name{Space: "", Local: "xmlns:xsd"}, Value: "http://www.w3.org/2001/XMLSchema"},
		{Name: xml.Name{Space: "", Local: "xmlns:soap"}, Value: "http://schemas.xmlsoap.org/soap/envelope/"},
	},
}

type baseContent struct {
	Body struct {
		Data interface{}
	} `xml:"soap:Body"`
	Header xml.Name `xml:"soap:Header"`
}

// MarshalXML envelope the body and encode to xml
func (c process) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	content := baseContent{
		Body: struct{ Data interface{} }{
			c.Request.Params,
		},
	}
	return e.EncodeElement(content, start)
}
