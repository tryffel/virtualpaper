package integrationtest

import "gopkg.in/h2non/baloo.v3"

const BASEURL = "http://localhost:8000"

type httpTest struct {
	client *baloo.Client
}

func (t *httpTest) Authorize() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Authorization", "Bearer "+UserToken),
	}
}

func (t *httpTest) AuthorizeAdmin() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Authorization", "Bearer "+AdminToken),
	}
}

func (t *httpTest) IsJson() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Content-Type", "application/json"),
	}
}

var client = &httpTest{client: baloo.New(BASEURL)}
