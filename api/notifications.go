package api

import "fmt"

const NotificationPath = "/notification"
const SubscribeNotificationPath = "/subscribe"
const SpecificNotificationPath = "/notification/%s"
const GetListenersPath = "/listening/%s"

const IdPathVar = "id"
const UsernamePathVar = "username"

var SpecificNotificationRoute = fmt.Sprintf(SpecificNotificationPath, fmt.Sprintf("{%s}", IdPathVar))
var GetListenersRoute = fmt.Sprintf(GetListenersPath, fmt.Sprintf("{%s}", UsernamePathVar))
