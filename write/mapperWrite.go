package write

import (
	"fmt"
	. "mybatis-generator/constant"
	"mybatis-generator/model"
	"mybatis-generator/util"
	"strings"
	"time"
)

type MapperWrite struct{}

// 数组里面的key均需要缩进/退回一个tab
var enterKeyboard = [...]string{
	"select",
	"update",
	"delete",
	"insert",
	"sql",
	"where",
	"resultMap",
	"mapper",
	"set",
	"trim",
}

//应该放在方法中传递,局部变量.单线程下无所谓.
var mapperTabNumber int
var primaryKeys []model.Field
var autoIncrement *model.Field

func (w *MapperWrite) Write(model *model.Model, ch chan string) {
	ch <- fmt.Sprintf("开始写入Mapper- %s", time.Now().String())
	file := util.CreateFile(fmt.Sprintf("%sMapper.xml", model.Name))
	defer file.Close()
	var content string

	writeStartDocumentType(&content, model.Name)

	writeResultMap(&content, model)
	writeTableName(&content, model.TableName)
	writeColsExcludeAutoIncrement(&content, model)
	writeColsAll(&content, model)
	writeCriteria(&content, model)
	writeCreate(&content, model)
	// writeDelete(&content, model)
	writeUpdate(&content, model)
	// writeFindById(&content, model)
	// writePaging(&content, model)
	writeEndDocumentType(&content)
	file.WriteString(content)
	ch <- fmt.Sprintf("写入Mapper完成- %s", time.Now().String())
}

func writeStartDocumentType(content *string, model string) {
	writeMapper(content, "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n<!-- \n ~ "+GeneratorInfo+time.Now().Format("2006/01/02 15:04:05")+".\n  -->\n<!DOCTYPE mapper PUBLIC \"-//mybatis.org//DTD Mapper 3.0//EN\" \"http://mybatis.org/dtd/mybatis-3-mapper.dtd\" >\n")
	writeMapper(content, "<mapper namespace=\""+model+"\">\n")
}

func writeResultMap(content *string, model *model.Model) {
	writeMapper(content, "<resultMap id=\""+model.Name+"Map\" type=\""+model.Name+"\">")
	for index, v := range *model.Fields {
		if strings.Contains(strings.ToLower(v.Extra.String), AutoIncrement) {
			autoIncrement = &(*model.Fields)[index]
		}
		if v.Key.String == "PRI" {
			writeMapper(content, "<id column=\""+v.Name+"\" property=\""+v.FieldName+"\"/>")
		} else {
			writeMapper(content, "<result column=\""+v.Name+"\" property=\""+v.FieldName+"\"/>")
		}
	}
	writeMapper(content, "</resultMap>")
}

func writeTableName(content *string, table string) {
	writeMapper(content, "<sql id=\"tb\">")
	writeMapper(content, "`"+table+"`")
	writeMapper(content, "</sql>")
}

func writeColsExcludeAutoIncrement(content *string, model *model.Model) {
	writeMapper(content, "<sql id=\"cols_exclude_id\">")
	s := []string{}
	for _, col := range *model.Fields {
		if strings.Contains(strings.ToLower(col.Extra.String), AutoIncrement) {
			continue
		}
		s = append(s, "`"+col.Name+"`")
	}
	writeMapper(content, strings.Join(s, ", "))
	writeMapper(content, "</sql>")
}

func writeColsAll(content *string, model *model.Model) {
	writeMapper(content, "<sql id=\"cols_all\">")
	s := []string{}
	for _, col := range *model.Fields {
		if !strings.Contains(strings.ToLower(col.Extra.String), AutoIncrement) {
			continue
		}
		s = append(s, "`"+col.Name+"`")
	}
	s = append(s, " <include refid=\"cols_exclude_id\"/>")

	writeMapper(content, strings.Join(s, ","))
	writeMapper(content, "</sql>")
}

func writeCriteria(content *string, model *model.Model) {
	writeMapper(content, "<sql id=\"criteria\">")
	writeMapper(content, "<where>")
	for _, v := range *model.Fields {
		writeMapper(content, "<if test=\""+v.FieldName+" != null\"> AND `"+v.Name+"` = #{"+v.FieldName+"}</if>")
	}
	writeMapper(content, "</where>")
	c := *content
	*content = c[0 : len(c)-1]
	writeMapper(content, "</sql>")
}

func writeCreate(content *string, m *model.Model) {
	startCreateSql := "<insert id=\"create\""
	if nil != autoIncrement {
		startCreateSql += " keyProperty=\"" + autoIncrement.Name + "\" useGeneratedKeys=\"true\""
	}
	startCreateSql += ">"
	writeMapper(content, startCreateSql)

	writeMapper(content, "INSERT INTO <include refid=\"tb\"/> (")
	writeMapper(content, "<trim suffix=\"\" suffixOverrides=\",\">")
	additionalFields := []model.Field{}
	for _, field := range *m.Fields {
		if strings.Contains(field.Extra.String, AutoIncrement) {
			continue
		}
		if strings.Contains(field.Comment.String, CreateTimeDesc) || strings.Contains(field.Comment.String, UpdateTimeDesc) {
			additionalFields = append(additionalFields, field)
		} else {
			writeMapper(content, "<if test=\""+field.FieldName+" != null\"> `"+field.Name+"`,</if>")
		}
	}
	additionalColumns := []string{}
	for _, field := range additionalFields {
		additionalColumns = append(additionalColumns, "`"+field.Name+"`")
	}
	writeMapper(content, strings.Join(additionalColumns, ","))
	writeMapper(content, "</trim>")
	writeMapper(content, ") VALUES (")
	writeMapper(content, "<trim suffix=\"\" suffixOverrides=\",\">")

	for _, field := range *m.Fields {
		if strings.Contains(field.Extra.String, AutoIncrement) {
			continue
		}
		if !strings.Contains(field.Comment.String, CreateTimeDesc) && !strings.Contains(field.Comment.String, UpdateTimeDesc) {
			writeMapper(content, "<if test=\""+field.FieldName+" != null\"> #{"+field.FieldName+"},</if>")
		}
	}
	additionalColumns = []string{}
	for _, _ = range additionalFields {
		additionalColumns = append(additionalColumns, "now()")
	}
	writeMapper(content, strings.Join(additionalColumns, ","))
	writeMapper(content, "</trim>")

	writeMapper(content, ")")
	writeMapper(content, "</insert>")
}

func writeUpdate(content *string, m *model.Model) {
	writeMapper(content, "<update id=\"update\">")
	writeMapper(content, "UPDATE <include refid=\"tb\"/> ")
	writeMapper(content, "<set>")
	var updateField *model.Field

	for index, field := range *m.Fields {
		if strings.Contains(field.Extra.String, AutoIncrement) {
			continue
		} else if strings.Contains(field.Comment.String, UpdateTimeDesc) {
			updateField = &(*m.Fields)[index]
		} else {
			writeMapper(content, "<if test=\""+field.FieldName+" != null\"> `"+field.Name+"` = #{"+field.FieldName+"}, </if>")
		}
	}
	if nil != updateField {
		writeMapper(content, updateField.Name+" = now()")
	}
	writeMapper(content, "</set>")

	if nil != autoIncrement {
		writeMapper(content, "WHERE "+autoIncrement.Name+" = #{"+autoIncrement.FieldName+"}")
	} else {
		for _, field := range *m.Fields {
			if strings.Contains(field.Key.String, PriKey) {
				writeMapper(content, " WHERE "+field.Name+" = #{"+field.FieldName+"}")
				break
			}
		}
	}
	writeMapper(content, "</update>")
}

func writeEndDocumentType(content *string) {
	writeMapper(content, "</mapper>")
}

func (w MapperWrite) writeInfo(f *string, info string) {

}

func writeMapper(f *string, info string) {
	if contains(info, BackSlant) {
		info += NewLine
		mapperTabNumber -= TabNum
	}

	info = strings.Join(make([]string, mapperTabNumber), Blank) + info
	util.AppendLine(f, info)
	if contains(info, Lt) {
		mapperTabNumber += TabNum
	}
}

func contains(info string, compare string) bool {
	for i := 0; i < len(enterKeyboard); i++ {
		if strings.Contains(info, compare+enterKeyboard[i]) {
			return true
		}
	}
	return false
}
