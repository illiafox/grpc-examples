package main

import (
	"encoding/base64"
	deliverypb "examples/01-proto/gen"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	person := &deliverypb.Person{
		Name:    "Nikita",
		Address: "Kyiv",
		ContactMethod: &deliverypb.Person_TelegramHandle{ // oneof, use wrapper type
			TelegramHandle: "@aboba",
		},
		CreatedAt: timestamppb.Now(),
		Packages: []*deliverypb.Package{
			{
				Id:                        1,
				Description:               "Smartphone",
				WeightKg:                  0.5,
				EstimatedDeliveryDuration: durationpb.New(48 * time.Hour),
			},
		},
		Metadata: map[string]string{
			"priority": "high",
			"source":   "mobile_app",
		},
	}

	data, err := proto.Marshal(person)
	if err != nil {
		log.Fatal("Failed to marshal:", err)
	}

	fmt.Println(base64.StdEncoding.EncodeToString(data))

	decoded := new(deliverypb.Person)
	if err = proto.Unmarshal(data, decoded); err != nil {
		log.Fatal("Failed to unmarshal:", err)
	}

	switch contact := decoded.ContactMethod.(type) {
	case *deliverypb.Person_TelegramHandle:
		fmt.Println("Telegram:", contact.TelegramHandle)
	case *deliverypb.Person_WhatsappNumber:
		fmt.Println("WhatsApp:", contact.WhatsappNumber)
	default:
		fmt.Println("No contact method provided")
	}

	spew.Dump(decoded)
}
