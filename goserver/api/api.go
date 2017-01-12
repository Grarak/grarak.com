package api

import "../miniserver"

type Api interface {
	GetResponse() *miniserver.Response
}
