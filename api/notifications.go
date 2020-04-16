package api

import "fmt"

const NotificationPath = "/notification"
const SubscribeNotificationPath = "/subscribe"
const UnsubscribeNotificationPath = "/unsubscribe"
const SpecificNotificationPath = "/notification/%s"
const GetListenersPath = "/listening"

const IdPathVar = "id"

var SpecificNotificationRoute = fmt.Sprintf(SpecificNotificationPath, fmt.Sprintf("{%s}", IdPathVar))
