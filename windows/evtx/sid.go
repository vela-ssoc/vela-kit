package evtx

import (
	"fmt"
	"strconv"
)

// SID represents the Windows Security Identifier for an account.
type SID struct {
	Identifier string `xml:"UserID,attr"`
	Name       string
	Domain     string
	Type       SIDType
}

// String returns string representation of SID.
func (a SID) String() string {
	return fmt.Sprintf("SID Identifier[%s] Name[%s] Domain[%s] Typ[%s]",
		a.Identifier, a.Name, a.Domain, a.Type)
}

// SIDType identifies the type of a security identifier (SID).
type SIDType uint32

// SIDType values.
const (
	// Do not reorder.
	SidTypeUser SIDType = 1 + iota
	SidTypeGroup
	SidTypeDomain
	SidTypeAlias
	SidTypeWellKnownGroup
	SidTypeDeletedAccount
	SidTypeInvalid
	SidTypeUnknown
	SidTypeComputer
	SidTypeLabel
	SidTypeLogonSession
)

// sidTypeToString is a mapping of SID types to their string representations.
var sidTypeToString = map[SIDType]string{
	SidTypeUser:           "User",
	SidTypeGroup:          "Group",
	SidTypeDomain:         "Domain",
	SidTypeAlias:          "Alias",
	SidTypeWellKnownGroup: "Well Known Group",
	SidTypeDeletedAccount: "Deleted Account",
	SidTypeInvalid:        "Invalid",
	SidTypeUnknown:        "Unknown",
	SidTypeComputer:       "Computer",
	SidTypeLabel:          "Label",
	SidTypeLogonSession:   "Logon Session",
}

// String returns string representation of SIDType.
func (st SIDType) String() string {
	if typ, found := sidTypeToString[st]; found {
		return typ
	} else if st > 0 {
		return strconv.FormatUint(uint64(st), 10)
	}
	return ""
}
