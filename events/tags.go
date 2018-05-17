package events

/*
In the context of our system there are five main players.
1. Local Event Generators
2. Room Event Proxies
3. External Event Translator
4. Local Event Consumers
5. Local Proxy (this microservice)


For the purposes of this microservice, all events flow through the local proxy.
Different event types have different routing rules to the different players in the system
*/

const (
	ROOM       = "room"
	UI         = "ui"
	APISuccess = "api-success"
	APIError   = "api-error"
	Translator = "translator"
	External   = "external"
	Health     = "health"
	Metrics    = "metrics"
	UIFeature  = "uifeature"
	RoomDivide = "roomdivide"
)

// for measuring status of event routing
const (
	TestStart         = "teststart"
	TestPleaseReply   = "testpleasereply"
	TestExternal      = "testexternal"
	TestExternalReply = "testexternalreply"
	TestReply         = "testreply"
	TestEnd           = "testend"
)
