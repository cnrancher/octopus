package physical

import (
	"fmt"
	"testing"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var text = []byte(`{ 
	"store": {
		"book": [ 
		  { "category": "reference",
			"author": "Nigel Rees",
			"title": "Sayings of the Century",
			"price": 8.95
		  },
		  { "category": "fiction",
			"author": "Evelyn Waugh",
			"title": "Sword of Honour",
			"price": 12.99
		  },
		  { "category": "fiction",
			"author": "Herman Melville",
			"title": "Moby Dick",
			"isbn": "0-553-21311-3",
			"price": 8.99
		  },
		  { "category": "fiction",
			"author": "J. R. R. Tolkien",
			"title": "The Lord of the Rings",
			"isbn": "0-395-19395-8",
			"price": 22.99
		  }
		],
		"bicycle": {
		  "color": "red",
		  "price": 19.95
		}
	  }
	}`)

var text2 = []byte(`{"switch":"off","brightness"：4,"power":{"powerDissipation":"10KWH","electricQuantity":19.99}}`)

func TestConvertToStatusProperty(t *testing.T) {
	var property v1alpha1.Property
	property.JSONPath = "store.bicycle.price"
	property.Name = "test_property"
	property.Description = "test property"
	property.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
	property.SubInfo.Topic = "test/abc"
	property.SubInfo.Qos = 2
	statusProperty, err := ConvertToStatusProperty(text, &property)
	if err != nil {
		fmt.Println("ConvertToStatusProperty error:", err)
		return
	}

	var device v1alpha1.MqttDevice
	device.APIVersion = "v1alpha1"
	device.Kind = "MqttDevice"
	device.Name = "testDevice"
	device.Spec.Properties = append(device.Spec.Properties, property)
	device.Status.Properties = append(device.Status.Properties, statusProperty)

	var out = unstructured.Unstructured{Object: make(map[string]interface{})}
	var scheme = k8sruntime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	scheme.Convert(&device, &out, nil)
	var bytes, _ = out.MarshalJSON()
	fmt.Println(string(bytes))

	var property2 v1alpha1.Property
	property2.JSONPath = "power"
	property2.Name = "test_property"
	property2.Description = "test property"
	property2.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
	property2.SubInfo.Topic = "test/abc"
	property2.SubInfo.Qos = 2
	statusProperty2, err := ConvertToStatusProperty(text2, &property2)
	if err != nil {
		fmt.Println("ConvertToStatusProperty error:", err)
		return
	}

	var device2 v1alpha1.MqttDevice
	device2.APIVersion = "v1alpha1"
	device2.Kind = "MqttDevice"
	device2.Name = "testDevice"
	device2.Spec.Properties = append(device2.Spec.Properties, property2)
	device2.Status.Properties = append(device2.Status.Properties, statusProperty2)

	var out2 = unstructured.Unstructured{Object: make(map[string]interface{})}
	var scheme2 = k8sruntime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme2))
	scheme2.Convert(&device2, &out2, nil)
	var bytes2, _ = out2.MarshalJSON()
	fmt.Println(string(bytes2))

}

func TestConvertValueToJSONPayload(t *testing.T) {
	var property v1alpha1.Property
	property.JSONPath = "store.bicycle.color"
	property.Name = "test_property"
	property.Description = "test property"
	property.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
	property.SubInfo.Topic = "test/abc"
	property.SubInfo.Qos = 2
	property.Value.ValueType = v1alpha1.ValueTypeString
	property.Value.StringValue = "huang"
	newPayload, err := ConvertValueToJSONPayload(text, &property)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(newPayload))

	var propertyFloat v1alpha1.Property
	propertyFloat.JSONPath = "store.bicycle.price"
	propertyFloat.Name = "test_property"
	propertyFloat.Description = "test property"
	propertyFloat.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
	propertyFloat.SubInfo.Topic = "test/abc"
	propertyFloat.SubInfo.Qos = 2
	propertyFloat.Value.ValueType = v1alpha1.ValueTypeFloat
	propertyFloat.Value.FloatValue = new(v1alpha1.ValueFloat)
	propertyFloat.Value.FloatValue.F = float64(20.11)
	newPayload2, err := ConvertValueToJSONPayload(text, &propertyFloat)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(newPayload2))

	var propertyObject v1alpha1.Property
	propertyObject.JSONPath = "store.bicycle"
	propertyObject.Name = "test_property"
	propertyObject.Description = "test property"
	propertyObject.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
	propertyObject.SubInfo.Topic = "test/abc"
	propertyObject.SubInfo.Qos = 2
	propertyObject.Value.ValueType = v1alpha1.ValueTypeObject
	propertyObject.Value.ObjectValue = new(k8sruntime.RawExtension)
	propertyObject.Value.ObjectValue.Raw = []byte(`{"color":"black","price":222.77}`)
	newPayload3, err := ConvertValueToJSONPayload(text, &propertyObject)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(newPayload3))
}
