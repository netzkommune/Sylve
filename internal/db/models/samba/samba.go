package sambaModels

type SambaSettings struct {
	ID                 int    `json:"id" gorm:"primaryKey"`
	UnixCharset        string `json:"unixCharset"`
	Workgroup          string `json:"workgroup"`
	ServerString       string `json:"serverString" gorm:"default:'Sylve SMB Server'"`
	Interfaces         string `json:"interfaces" gorm:"default:'lo0'"`
	BindInterfacesOnly bool   `json:"bindInterfacesOnly" gorm:"default:true"`
}
