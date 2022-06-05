package module

import "github.com/allentom/harukap/module/notification"

var Notification *notification.NotificationModule

func CreateNotificationModule() error {
	Notification = &notification.NotificationModule{}
	err := Notification.InitModule()
	if err != nil {
		return err
	}
	return nil
}
