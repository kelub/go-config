package mysql

import (
	"time"
)

type Config struct {
	Id        int64  `xorm:"pk autoincr BIGINT(20)"`
	Key       string `json:"key"         xorm:"VARCHAR(128) index not null comment('配置键')"`
	Value     string `json:"value"       xorm:"VARCHAR(128) not null comment('配置值')"`
	Lastindex int64  `json:"lastindex"   xorm:"BIGINT(20) not null comment('最后修改标示')"`
}

type ConfigHistory struct {
	Id         int64     `xorm:"pk autoincr BIGINT(20)"`
	Key        string    `json:"key"          xorm:"VARCHAR(128) index not null comment('配置键')"`
	Value      string    `json:"value"        xorm:"VARCHAR(128) not null comment('配置值')"`
	Lastindex  int64     `json:"lastindex"    xorm:"BIGINT(20) not null comment('最后修改标示')"`
	Createtime time.Time `json:"createtime"   xorm:"TIMESTAMP default 'CURRENT_TIMESTAMP' comment('创建时间')"`
}
