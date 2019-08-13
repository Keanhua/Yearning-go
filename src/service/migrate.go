package service

import (
	"Yearning-go/src/lib"
	"Yearning-go/src/modal"
	"Yearning-go/src/parser"
	"encoding/json"
	"fmt"
)

func JsonCreate(o *parser.AuditRole, other *modal.Other, ldap *modal.Ldap, message *modal.Message, a *modal.PermissionList) {
	c, _ := json.Marshal(o)
	oh, _ := json.Marshal(other)
	l, _ := json.Marshal(ldap)
	m, _ := json.Marshal(message)
	ak, _ := json.Marshal(a)
	modal.DB().Create(&modal.CoreGlobalConfiguration{
		Authorization: "global",
		Other:         oh,
		AuditRole:     c,
		Message:       m,
		Ldap:          l,
	})
	modal.DB().Create(&modal.CoreGrained{
		Username:    "admin",
		Permissions: ak,
	})
}

func Migrate() {
	if !modal.DB().HasTable("core_accounts") {
		modal.DB().CreateTable(&modal.CoreAccount{})
		modal.DB().CreateTable(&modal.CoreDataSource{})
		modal.DB().CreateTable(&modal.CoreGlobalConfiguration{})
		modal.DB().CreateTable(&modal.CoreGrained{})
		modal.DB().CreateTable(&modal.CoreSqlOrder{})
		modal.DB().CreateTable(&modal.CoreSqlRecord{})
		modal.DB().CreateTable(&modal.CoreRollback{})
		modal.DB().CreateTable(&modal.CoreQueryRecord{})
		modal.DB().CreateTable(&modal.CoreQueryOrder{})
		modal.DB().CreateTable(&modal.CoreGroupOrder{})

		o := parser.AuditRole{
			DMLInsertColumns:               false,
			DMLMaxInsertRows:               10,
			DMLWhere:                       false,
			DMLOrder:                       false,
			DMLSelect:                      false,
			DDLCheckTableComment:           false,
			DDLCheckColumnNullable:         false,
			DDLCheckColumnDefault:          false,
			DDLEnableAcrossDBRename:        false,
			DDLEnableAutoincrementInit:     false,
			DDLEnableAutoincrementUnsigned: false,
			DDLEnableDropTable:             false,
			DDLEnableDropDatabase:          false,
			DDLEnableNullIndexName:         false,
			DDLIndexNameSpec:               false,
			DDLMaxKeyParts:                 5,
			DDLMaxKey:                      5,
			DDLMaxCharLength:               10,
			MaxTableNameLen:                10,
			MaxAffectRows:                  1000,
			EnableSetCollation:             false,
			EnableSetCharset:               false,
			SupportCharset:                 "",
			SupportCollation:               "",
			CheckIdentifier:                false,
			MustHaveColumns:                "",
			DDLMultiToSubmit:               false,
			OscAlterForeignKeysMethod:      "rebuild_constraints",
			OscMaxLag:                      1,
			OscChunkTime:                   0.5,
			OscMaxThreadConnected:          25,
			OscMaxThreadRunning:            25,
			OscCriticalThreadConnected:     20,
			OscCriticalThreadRunning:       20,
			OscRecursionMethod:             "processlist",
			OscCheckInterval:               1,
		}

		other := modal.Other{
			Limit:            "1000",
			IDC:              []string{"Aliyun", "AWS"},
			Multi:            false,
			Query:            false,
			ExcludeDbList:    []string{},
			InsulateWordList: []string{},
			Register:         false,
			Export:           false,
			ExQueryTime:      60,
			PerOrder:         2,
		}

		ldap := modal.Ldap{
			Url:      "",
			User:     "",
			Password: "",
			Type:     1,
			Sc:       "",
		}

		message := modal.Message{
			WebHook:  "",
			Host:     "",
			Port:     25,
			User:     "",
			Password: "",
			ToUser:   "",
			Mail:     false,
			Ding:     false,
			Ssl:      false,
		}

		a := modal.PermissionList{
			DDL:         "1",
			DML:         "1",
			Query:       "1",
			DDLSource:   []string{},
			DMLSource:   []string{},
			QuerySource: []string{},
			Auditor:     []string{},
			User:        "1",
			Base:        "1",
		}

		JsonCreate(&o, &other, &ldap, &message, &a)

		modal.DB().Create(&modal.CoreAccount{
			Username:   "admin",
			RealName:   "超级管理员",
			Password:   lib.DjangoEncrypt("Yearning_admin", string(lib.GetRandom())),
			Rule:       "admin",
			Department: "DBA",
			Email:      "",
		})

		fmt.Println("初始化成功!\n 用户名: admin\n密码:Yearning_admin")
	} else {
		fmt.Println("已初始化过,请不要再次执行")
	}
}

func UpdateSoft() {
	modal.DB().AutoMigrate(&modal.CoreAccount{})
	modal.DB().AutoMigrate(&modal.CoreDataSource{})
	modal.DB().AutoMigrate(&modal.CoreGlobalConfiguration{})
	modal.DB().AutoMigrate(&modal.CoreGrained{})
	modal.DB().AutoMigrate(&modal.CoreSqlOrder{})
	modal.DB().AutoMigrate(&modal.CoreSqlRecord{})
	modal.DB().AutoMigrate(&modal.CoreRollback{})
	modal.DB().AutoMigrate(&modal.CoreQueryRecord{})
	modal.DB().Debug().AutoMigrate(&modal.CoreQueryOrder{})
	modal.DB().AutoMigrate(&modal.CoreGroupOrder{})
	fmt.Println("数据已更新!")
}
