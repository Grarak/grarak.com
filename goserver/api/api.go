package api

import "../miniserver"

type Interface interface {
        GetResponse() *miniserver.Response
}
