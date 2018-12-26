package hubrpc

import (
	"fmt"
	"github.com/bitlum/hub/lightning"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func convertProtoMessage(resp proto.Message) string {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
		OrigName:     true,
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		return fmt.Sprintf("unable to decode response: %v", err)
	}

	return jsonStr
}

func convertPaymentStatusToProto(status lightning.PaymentStatus) (PaymentStatus, error) {
	var protoStatus PaymentStatus
	switch status {
	case lightning.Waiting:
		protoStatus = PaymentStatus_WAITING
	case lightning.Completed:
		protoStatus = PaymentStatus_COMPLETED
	case lightning.Pending:
		protoStatus = PaymentStatus_PENDING
	case lightning.Failed:
		protoStatus = PaymentStatus_FAILED
	default:
		protoStatus = PaymentStatus_STATUS_NONE
	}

	return protoStatus, nil
}

func convertPaymentDirectionToProto(direction lightning.PaymentDirection) (PaymentDirection,
	error) {
	var protoDirection PaymentDirection
	switch direction {
	case lightning.Outgoing:
		protoDirection = PaymentDirection_OUTGOING
	case lightning.Incoming:
		protoDirection = PaymentDirection_INCOMING
	default:
		protoDirection = PaymentDirection_DIRECTION_NONE
	}

	return protoDirection, nil
}

func convertPaymentSystemToProto(system lightning.PaymentSystem) (
	PaymentSystem, error) {
	var protoSystem PaymentSystem
	switch system {
	case lightning.Internal:
		protoSystem = PaymentSystem_INTERNAL
	case lightning.External:
		protoSystem = PaymentSystem_EXTERNAL
	default:
		protoSystem = PaymentSystem_SYSTEM_NONE
	}

	return protoSystem, nil
}

func convertPaymentToProto(payment *lightning.Payment) (*Payment, error) {
	status, err := convertPaymentStatusToProto(payment.Status)
	if err != nil {
		return nil, err
	}

	direction, err := convertPaymentDirectionToProto(payment.Direction)
	if err != nil {
		return nil, err
	}

	system, err := convertPaymentSystemToProto(payment.System)
	if err != nil {
		return nil, err
	}

	return &Payment{
		PaymentId:   payment.PaymentID,
		UpdatedAt:   payment.UpdatedAt,
		Status:      status,
		Direction:   direction,
		System:      system,
		PaymentHash: string(payment.PaymentHash),
		Amount:      payment.Amount.String(),
		MediaFee:    payment.MediaFee.String(),
	}, nil
}

func ConvertPaymentStatusFromProto(protoStatus PaymentStatus) (
	lightning.PaymentStatus, error) {
	var status lightning.PaymentStatus
	switch protoStatus {
	case PaymentStatus_WAITING:
		status = lightning.Waiting
	case PaymentStatus_COMPLETED:
		status = lightning.Completed
	case PaymentStatus_PENDING:
		status = lightning.Pending
	case PaymentStatus_FAILED:
		status = lightning.Failed
	case PaymentStatus_STATUS_NONE:
		status = ""
	default:
		return status, errors.Errorf("unable convert unknown status: %v",
			protoStatus)
	}

	return status, nil
}

func ConvertPaymentDirectionFromProto(protoDirection PaymentDirection) (
	lightning.PaymentDirection, error) {
	var direction lightning.PaymentDirection
	switch protoDirection {
	case PaymentDirection_OUTGOING:
		direction = lightning.Outgoing
	case PaymentDirection_INCOMING:
		direction = lightning.Incoming
	case PaymentDirection_DIRECTION_NONE:
		direction = ""
	default:
		return direction, errors.Errorf("unable convert unknown direction: %v",
			protoDirection)
	}

	return direction, nil
}

func ConvertPaymentSystemFromProto(protoSystem PaymentSystem) (
	lightning.PaymentSystem, error) {
	var system lightning.PaymentSystem
	switch protoSystem {
	case PaymentSystem_INTERNAL:
		system = lightning.Internal
	case PaymentSystem_EXTERNAL:
		system = lightning.External
	case PaymentSystem_SYSTEM_NONE:
		system = ""
	default:
		return system, errors.Errorf("unable convert unknown system: %v",
			protoSystem)
	}

	return system, nil
}
