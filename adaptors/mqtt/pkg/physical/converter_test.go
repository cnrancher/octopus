package physical

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var simpleData = []byte(`{ 
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

var roomLightData = []byte(`{"switch":"off","brightness"：4,"power":{"powerDissipation":"10KWH","electricQuantity":19.99}}`)

func TestConvertToStatusProperty(t *testing.T) {

	tests := map[string]struct {
		input    []byte
		jsonPath string
		want     []byte
	}{
		"simple": {
			input:    simpleData,
			jsonPath: "store.bicycle.price",
			want:     []byte(`{"apiVersion":"devices.edge.cattle.io/v1alpha1","kind":"MqttDevice","metadata":{"creationTimestamp":null,"name":"testDevice"},"spec":{"config":{"broker":""},"properties":[{"description":"test property","jsonPath":"store.bicycle.price","name":"test_property","pubInfo":{"qos":0,"topic":""},"subInfo":{"payloadType":"json","qos":2,"topic":"test/abc"},"value":{"valueType":""}}]},"status":{"properties":[{"description":"test property","name":"test_property","updateAt":"2020-01-01T01:01:01Z","value":{"floatValue":"19.950000","valueType":"float"}}]}}`),
		},
		"room light": {
			input:    roomLightData,
			jsonPath: "power",
			want:     []byte(`{"apiVersion":"devices.edge.cattle.io/v1alpha1","kind":"MqttDevice","metadata":{"creationTimestamp":null,"name":"testDevice"},"spec":{"config":{"broker":""},"properties":[{"description":"test property","jsonPath":"power","name":"test_property","pubInfo":{"qos":0,"topic":""},"subInfo":{"payloadType":"json","qos":2,"topic":"test/abc"},"value":{"valueType":""}}]},"status":{"properties":[{"description":"test property","name":"test_property","updateAt":"2020-01-01T01:01:01Z","value":{"objectValue":{"electricQuantity":19.99,"powerDissipation":"10KWH"},"valueType":"object"}}]}}`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var property v1alpha1.Property
			property.JSONPath = tc.jsonPath
			property.Name = "test_property"
			property.Description = "test property"
			property.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
			property.SubInfo.Topic = "test/abc"
			property.SubInfo.Qos = 2
			statusProperty, err := ConvertToStatusProperty(tc.input, &property)
			if err != nil {
				t.Fatal("ConvertToStatusProperty error:", err)
				return
			}
			tm := time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC)
			statusProperty.UpdatedAt = metav1.Time{Time: tm}
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
			got := bytes[:len(bytes)-1]

			if !reflect.DeepEqual(got, tc.want) {
				log.Fatalf("expected: %s , got: %s", string(tc.want), string(bytes))
			}
		})
	}
}

func TestConvertValueToJSONPayload(t *testing.T) {

	tests := map[string]struct {
		input    []byte
		jsonPath string
		want     []byte
	}{
		"simple string": {
			input:    simpleData,
			jsonPath: "store.bicycle.color",
			want: []byte(`{ 
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
		  "color": "huang",
		  "price": 19.95
		}
	  }
	}`),
		},
		"simple float": {
			input:    simpleData,
			jsonPath: "store.bicycle.price",
			want: []byte(`{ 
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
		  "price": 20.11
		}
	  }
	}`),
		},
		"simple array": {
			input:    simpleData,
			jsonPath: "store.book",
			want: []byte(`{ 
	"store": {
		"book": [10,30],
		"bicycle": {
		  "color": "red",
		  "price": 19.95
		}
	  }
	}`),
		},
		"simple object": {
			input:    simpleData,
			jsonPath: "store.bicycle",
			want: []byte(`{ 
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
		"bicycle": {"color":"black","price":222.77}
	  }
	}`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var property v1alpha1.Property
			property.JSONPath = tc.jsonPath
			property.Name = "test_property"
			property.Description = "test property"
			property.SubInfo.PayloadType = v1alpha1.PayloadTypeJSON
			property.SubInfo.Topic = "test/abc"
			property.SubInfo.Qos = 2

			switch name {
			case "simple string":
				property.Value.ValueType = v1alpha1.ValueTypeString
				property.Value.StringValue = "huang"
			case "simple float":
				property.Value.ValueType = v1alpha1.ValueTypeFloat
				property.Value.FloatValue = new(v1alpha1.ValueFloat)
				property.Value.FloatValue.F = float64(20.11)
			case "simple array":
				property.Value.ValueType = v1alpha1.ValueTypeArray
				item1 := v1alpha1.ValueArrayProps{ValueProps: v1alpha1.ValueProps{ValueType: v1alpha1.ValueTypeInt, IntValue: 10}}
				item2 := v1alpha1.ValueArrayProps{ValueProps: v1alpha1.ValueProps{ValueType: v1alpha1.ValueTypeInt, IntValue: 30}}
				property.Value.ArrayValue = append(property.Value.ArrayValue, item1, item2)
			case "simple object":
				property.Value.ValueType = v1alpha1.ValueTypeObject
				property.Value.ObjectValue = new(k8sruntime.RawExtension)
				property.Value.ObjectValue.Raw = []byte(`{"color":"black","price":222.77}`)
			}

			newPayload, err := ConvertValueToJSONPayload(simpleData, &property)
			if err != nil {
				t.Fatal(err)
				return
			}

			if !reflect.DeepEqual(newPayload, tc.want) {
				log.Fatalf("expected: %s , got: %s", string(tc.want), string(newPayload))
			}
		})
	}
}
