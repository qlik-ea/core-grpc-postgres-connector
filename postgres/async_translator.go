package postgres

import (
	"github.com/jackc/pgx"
	"github.com/qlik-ea/postgres-grpc-connector/qlik"
)

func getTypeConstants(fieldDescriptors []pgx.FieldDescription) []qlik.FieldType {
	var fieldTypes = make([]qlik.FieldType, len(fieldDescriptors))
	for i, fieldDescr := range fieldDescriptors {
		switch fieldDescr.DataTypeName {
		case "int4":
			fieldTypes[i] = qlik.FieldType_INTEGER
		default:
			fieldTypes[i] = qlik.FieldType_TEXT
		}
	}
	return fieldTypes
}

/**
 *	Class AsyncStreamWriter
 */

type AsyncTranslator struct {
	writer *qlik.AsyncStreamWriter
	fieldDescriptors []pgx.FieldDescription
	channel chan [][]interface{}
}

func NewAsyncTranslator(writer *qlik.AsyncStreamWriter, fieldDescriptors []pgx.FieldDescription) *AsyncTranslator {
	var this = &AsyncTranslator{writer, fieldDescriptors, make(chan [][]interface{}, 10)}
	go this.run()
	return this
}

func ( this *AsyncTranslator) GetDataResponseMetadata() *qlik.GetDataResponse {
	var types = getTypeConstants(this.fieldDescriptors)
	var array = make([]*qlik.FieldInfo, len(this.fieldDescriptors))

	for i := range this.fieldDescriptors {
		array[i] = &qlik.FieldInfo{this.fieldDescriptors[i].Name, types[i], []qlik.FieldTag{}}
	}
	return &qlik.GetDataResponse{array, ""}
}

func ( this *AsyncTranslator) buildRowBundle(tempQixRowList [][]interface{}) *qlik.BundledRows {
	var typeConsts = getTypeConstants(this.fieldDescriptors)
	var columnCount, rowCount = len(this.fieldDescriptors), int64(len(tempQixRowList))
	var rowBundle = qlik.BundledRows{Cols: make([]*qlik.Column, columnCount)}

	for i := 0; i < columnCount; i++ {
		var column = &qlik.Column{}
		switch typeConsts[i] {
		case qlik.FieldType_TEXT:
			column.Flags=make([]qlik.ValueFlag, rowCount)
			column.Strings=make([]string, rowCount)
			column.Numbers=nil
			for r := 0; r < len(tempQixRowList); r++ {
				var srcValue = tempQixRowList[r][i]
				if srcValue != nil {
					column.Strings[r] = srcValue.(string)
					column.Flags[r] = qlik.ValueFlag_Normal
				} else {
					column.Strings[r] = ""
					column.Flags[r] = qlik.ValueFlag_Null
				}
			}
		case qlik.FieldType_INTEGER:
			column.Flags = nil
			column.Strings=nil
			column.Numbers=make([]float64, rowCount)
			for r := 0; r < len(tempQixRowList); r++ {
				var srcValue = tempQixRowList[r][i]
				column.Numbers[r] = float64(int64(srcValue.(int32)))
			}
		}
		rowBundle.Cols[i] = column
	}
	return &rowBundle
}

func (this *AsyncTranslator) Write(values [][]interface{}) {
	this.channel <- values
}

func (this *AsyncTranslator) Close() {
	close(this.channel)
}

func (this *AsyncTranslator) run() {
	for tempQixRowList := range this.channel {
		var resultChunk = this.buildRowBundle(tempQixRowList)
		this.writer.Write(resultChunk)
	}
	this.writer.Close()
}