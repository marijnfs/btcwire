// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcwire

import (
	"fmt"
	"io"
)

// MsgInv implements the Message interface and represents a bitcoin inv message.
// It is used to advertise a peer's known data such as blocks and transactions
// through inventory vectors.  It may be sent unsolicited to inform other peers
// of the data or in response to a getblocks message (MsgGetBlocks).  Each
// message is limited to a maximum number of inventory vectors, which is
// currently 50,000.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending an inv message to another peer.
type MsgInv struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgInv) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgInv.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgInv) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgInv.BtcDecode", str)
	}

	for i := uint64(0); i < count; i++ {
		iv := InvVect{}
		err := readInvVect(r, pver, &iv)
		if err != nil {
			return err
		}
		msg.AddInvVect(&iv)
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgInv) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgInv.BtcEncode", str)
	}

	err := writeVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, iv := range msg.InvList {
		err := writeInvVect(w, pver, iv)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgInv) Command() string {
	return cmdInv
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgInv) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return maxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload)
}

// NewMsgInv returns a new bitcoin inv message that conforms to the Message
// interface.  See MsgInv for details.
func NewMsgInv() *MsgInv {
	return &MsgInv{}
}
