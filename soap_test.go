package gosoap

import (
	"encoding/xml"
	"net/http"
	"testing"
)

var (
	scts = []struct {
		URL string
		Err bool
	}{
		{
			URL: "://www.server",
			Err: false,
		},
		{
			URL: "",
			Err: false,
		},
		{
			URL: "http://ec.europa.eu/taxation_customs/vies/checkVatService.wsdl",
			Err: true,
		},
	}
)

func TestSoapClient(t *testing.T) {
	for _, sct := range scts {
		_, err := SoapClient(sct.URL)
		if err != nil && sct.Err {
			t.Errorf("URL: %s - error: %s", sct.URL, err)
		}
	}
}

type CheckVatRequest struct {
	CountryCode string
	VatNumber   string
}

type CapitalCity struct {
	XMLName         xml.Name `xml:"CapitalCity"`
	SCountryISOCode string   `xml:"sCountryISOCode"`
}

type Whois struct {
	DomainName string `xml:"DomainName"`
}

func (r CheckVatRequest) SoapBuildRequest() *Request {
	return NewRequest("checkVat", map[string]string{
		"countryCode": r.CountryCode,
		"vatNumber":   r.VatNumber,
	})
}

type CheckVatResponse struct {
	CountryCode string `xml:"countryCode"`
	VatNumber   string `xml:"vatNumber"`
	RequestDate string `xml:"requestDate"`
	Valid       string `xml:"valid"`
	Name        string `xml:"name"`
	Address     string `xml:"address"`
}

type CapitalCityResponse struct {
	CapitalCityResult string
}

type NumberToWordsResponse struct {
	NumberToWordsResult string
}

type WhoisResponse struct {
	WhoisResult string
}

type ReqParam struct {
	VatNumber   string
	CountryCode string
}

var (
	rv CheckVatResponse
	rc CapitalCityResponse
	rn NumberToWordsResponse
	rw WhoisResponse

	params = ReqParam{}
)

func TestClient_Call(t *testing.T) {

	soap, err := SoapClient("http://webservices.oorsprong.org/websamples.countryinfo/CountryInfoService.wso?WSDL")
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	res, err := soap.Call("CapitalCity", CapitalCity{SCountryISOCode: "GB"})
	if err != nil {
		t.Errorf("error in soap call: %s", err)
	}

	res.Unmarshal(&rc)

	if rc.CapitalCityResult != "London" {
		t.Errorf("error: %+v", rc)
	}

	soap, err = SoapClient("http://www.dataaccess.com/webservicesserver/numberconversion.wso?WSDL")
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	res, err = soap.Call("NumberToWords", map[string]string{"ubiNum": "23"})
	if err != nil {
		t.Errorf("error in soap call: %s", err)
	}

	res.Unmarshal(&rn)

	if rn.NumberToWordsResult != "twenty three " {
		t.Errorf("error: %+v", rn)
	}

	soap, err = SoapClient("https://domains.livedns.co.il/API/DomainsAPI.asmx?WSDL")
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	res, err = soap.Call("Whois", Whois{DomainName: "google.com"})
	if err != nil {
		t.Errorf("error in soap call: %s", err)
	}

	res.Unmarshal(&rw)

	if rw.WhoisResult != "0" {
		t.Errorf("error: %+v", rw)
	}

	c := &Client{}
	res, err = c.Call("", map[string]string{})
	if err == nil {
		t.Errorf("error expected but nothing got.")
	}

	c.SetWSDL("://test.")

	res, err = c.Call("checkVat", params)
	if err == nil {
		t.Errorf("invalid WSDL")
	}
}

func TestClient_CallByStruct(t *testing.T) {
	soap, err := SoapClient("http://ec.europa.eu/taxation_customs/vies/checkVatService.wsdl")
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	var res *Response

	res, err = soap.CallByStruct(CheckVatRequest{
		CountryCode: "IE",
		VatNumber:   "6388047V",
	})
	if err != nil {
		t.Errorf("error in soap call: %s", err)
	}

	res.Unmarshal(&rv)
	if rv.CountryCode != "IE" {
		t.Errorf("error: %+v", rv)
	}
}

func TestClient_Call_NonUtf8(t *testing.T) {
	soap, err := SoapClient("https://demo.ilias.de/webservice/soap/server.php?wsdl")
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	_, err = soap.Call("login", map[string]string{"client": "demo", "username": "robert", "password": "iliasdemo"})
	if err != nil {
		t.Errorf("error in soap call: %s", err)
	}
}

func TestProcess_doRequest(t *testing.T) {
	c := &process{
		Client: &Client{
			HttpClient: &http.Client{},
		},
	}

	_, err := c.doRequest("")
	if err == nil {
		t.Errorf("body is empty")
	}

	_, err = c.doRequest("://teste.")
	if err == nil {
		t.Errorf("invalid WSDL")
	}
}

func TestUseDefinitionURL(t *testing.T) {
	initDefinition := func() *wsdlDefinitions {
		soapAddresses := []*soapAddress{
			{Location: "http://demo.ilias.de/webservice/soap/server.php"},
		}
		ports := []*wsdlPort{
			{SoapAddresses: soapAddresses},
		}
		service := []*wsdlService{
			{Ports: ports},
		}
		definition := &wsdlDefinitions{Services: service}

		return definition
	}

	client := &Client{
		wsdl:             "https://demo.ilias.de:4330/webservice/soap/server.php?wsdl",
		HttpClient:       &http.Client{},
		UseDefinitionURL: true,
		Definitions:      initDefinition(),
	}

	if client.getLocation() != "https://demo.ilias.de:4330/webservice/soap/server.php" {
		t.Errorf("url invalid")
	}

	client2 := &Client{
		wsdl:             "https://demo.ilias.de:4330/webservice/soap/server.php?wsdl",
		HttpClient:       &http.Client{},
		UseDefinitionURL: false,
		Definitions:      initDefinition(),
	}

	if client2.getLocation() != "http://demo.ilias.de/webservice/soap/server.php" {
		t.Errorf("url invalid")
	}
}
