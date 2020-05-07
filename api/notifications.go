package api

import "fmt"

const NotificationPath = "/notifications"
const SubscribeNotificationPath = "/notifications/subscribe"
const UnsubscribeNotificationPath = "/notifications/unsubscribe"
const SpecificNotificationPath = "/notifications/notification/%s"
const GetListenersPath = "/notifications/listening"

const IdPathVar = "id"

var SpecificNotificationRoute = fmt.Sprintf(SpecificNotificationPath, fmt.Sprintf("{%s}", IdPathVar))
