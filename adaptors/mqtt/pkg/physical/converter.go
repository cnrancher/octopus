package physical

import (
	"math"
	"time"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const minInt53 = -2251799813685248
const maxInt53 = 2251799813685247
const diffMin = 0.000001

func ConvertToStatusProperty(payload []byte, property *v1alpha1.Property) (v1alpha1.StatusProperty, error) {
	var statusProperty v1alpha1.StatusProperty
	statusProperty.Description = property.Description
	statusProperty.Name = property.Name
	statusProperty.UpdatedAt = metav1.Time{Time: time.Now()}

	if property.SubInfo.PayloadType != v1alpha1.PayloadTypeJSON {
		return statusProperty, nil
	}

	result := gjson.GetBytes(payload, property.JSONPath)
	valueProps, err := ConvertResultToValueProps(result)
	if err != nil {
		return statusProperty, err
	}
	statusProperty.Value = valueProps

	return statusProperty, nil
}

func ConvertResultToValueProps(result gjson.Result) (v1alpha1.ValueProps, error) {
	var valueProps v1alpha1.ValueProps
	var err error
	switch result.Type {
	case gjson.Null:
		return valueProps, nil
	case gjson.False, gjson.True:
		valueProps.ValueType = v1alpha1.ValueTypeBoolean
		boolValue := result.Bool()
		valueProps.BooleanValue = boolValue
	case gjson.String:
		valueProps.ValueType = v1alpha1.ValueTypeString
		stringValue := result.String()
		valueProps.StringValue = stringValue
	case gjson.JSON:
		if result.IsArray() {
			valueProps.ValueType = v1alpha1.ValueTypeArray
			valueProps.ArrayValue = new(runtime.RawExtension)
			valueProps.ArrayValue.Raw = []byte(result.String())
		} else if result.IsObject() {
			valueProps.ValueType = v1alpha1.ValueTypeObject
			valueProps.ObjectValue = new(runtime.RawExtension)
			valueProps.ObjectValue.Raw = []byte(result.String())
		}
	case gjson.Number:
		n := int64(result.Num)
		if float64(n) == result.Num && n >= minInt53 && n <= maxInt53 {
			valueProps.ValueType = v1alpha1.ValueTypeInt
			valueProps.IntValue = n
		} else {
			valueProps.ValueType = v1alpha1.ValueTypeFloat
			valueProps.FloatValue = new(v1alpha1.ValueFloat)
			valueProps.FloatValue.F = result.Num
		}
	}
	return valueProps, err
}

func ConvertValueToJSONPayload(payload []byte, property *v1alpha1.Property) ([]byte, error) {

	var value interface{}
	var err error
	switch property.Value.ValueType {
	case v1alpha1.ValueTypeArray:
		value = property.Value.ArrayValue
	case v1alpha1.ValueTypeObject:
		value = property.Value.ObjectValue
	case v1alpha1.ValueTypeInt:
		value = property.Value.IntValue
	case v1alpha1.ValueTypeString:
		value = property.Value.StringValue
	case v1alpha1.ValueTypeBoolean:
		value = property.Value.BooleanValue
	case v1alpha1.ValueTypeFloat:
		value = property.Value.FloatValue.F
	}

	nValue, err := sjson.SetBytes(payload, property.JSONPath, value)

	if err != nil {
		return nil, err
	}

	return nValue, nil

}

func ComparativeValueProps(l, r v1alpha1.ValueProps) bool {
	switch l.ValueType {
	case v1alpha1.ValueTypeArray:
		ls := string(l.ArrayValue.Raw)
		rs := string(r.ArrayValue.Raw)
		if ls != rs {
			return false
		}
	case v1alpha1.ValueTypeObject:
		ls := string(l.ObjectValue.Raw)
		rs := string(r.ObjectValue.Raw)
		if ls != rs {
			return false
		}
	case v1alpha1.ValueTypeInt:
		if l.IntValue != r.IntValue {
			return false
		}
	case v1alpha1.ValueTypeString:
		if l.StringValue != r.StringValue {
			return false
		}
	case v1alpha1.ValueTypeBoolean:
		if l.BooleanValue != r.BooleanValue {
			return false
		}
	case v1alpha1.ValueTypeFloat:
		if math.Dim(l.FloatValue.F, r.FloatValue.F) < diffMin {
			return false
		}
	}

	return true
}
