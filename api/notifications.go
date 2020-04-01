package api

import "fmt"

const NotificationPath = "/notification"
const SubscribeNotificationPath = "/subscribe"
const SpecificNotificationPath = "/notification/%s"

const IdPathVar = "id"

var SpecificNotificationRoute = fmt.Sprintf(SpecificNotificationPath, fmt.Sprintf("{%s}", IdPathVar))
