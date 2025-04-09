package data

import "time"

// сообщения от активного клиента во время синхронизации с активируемым клиентом
const (
	WRONG_PASS = "wrong password"
	NOT_APPROVED = "new client not approved"
	APPROVED = "new client approved"
)

// IpInfo описывает ответ сервиса ipinfo.io
type IpInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"` // Координаты в формате "широта,долгота"
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

// ClientData данные хранимые только на клиенте
type ClientData struct {
	SaltPW string
	Err error
}

// ClientInfo информация о клиенте
type ClientInfo struct {
	OSname string // операционная система
	BaseSerial string // серийник материнской платы
	IP string 
	MAC string // мак адрес
	Local time.Location // часовой пояс
	Location IpInfo // инфа о локации, которая получаем с помощью сайта
	TempID string // временный id генерируемый на стороне нового пользователя
	Uname string // имя пользователя, введённая на новом клиенте
	PW string // пароль введённый на клиенте
}