package main

import (
	"slices"
	"strings"

	authpb "github.com/nova38/saacs/pkg/chaincode/gen/auth/v1"
	"github.com/samber/oops"
)

const sep = string(rune(0))

// MakeComposeKey creates a composite key from the given attributes
//
// Based on the "github.com/hyperledger/fabric-chaincode-go/shim" package's CreateCompositeKey
// so that we don't have to import the shim package in this package
func MakeComposeKey(namespace string, attrs []string) (key string, err error) {
	// return shim.CreateCompositeKey(namespace, []string{attrs})
	key = namespace + sep

	for _, attr := range attrs {
		// TODO Validate the attribute

		key = key + attr + sep
	}
	return key, nil

}

// ─────────────────────────────────────────────────────────────────────────────────
// ─────────────────────────────────────────────────────────────────────────────────

// MakeStateKey creates a composite key from the given attributes
// Key should be {<ITEM_TYPE>}{COLLECTION_ID}{...ITEM_ID}
// Panics if ItemType or CollectionId is nil or an empty string
func MakeStateKey(objKey *authpb.ItemKey) (key string) {

	attrs := objKey.GetItemKeyParts()
	if attrs == nil {
		panic("ItemKeyParts is nil")
	}

	collectionId := objKey.GetCollectionId()
	if collectionId == "" {
		panic("CollectionId is nil")
	}

	key = sep + objKey.GetItemType() + sep + collectionId + sep

	if len(attrs) == 0 {
		return key
	}
	key = key + strings.Join(attrs, sep) + sep

	return key
}

func KeyToSubKey(objKey *authpb.ItemKey, subType string) (subKey *authpb.ItemKey) {
	subKey = &authpb.ItemKey{
		ItemType:     subType,
		CollectionId: objKey.GetCollectionId(),
		ItemKeyParts: objKey.GetItemKeyParts(),
	}

	return subKey
}

// ─────────────────────────────────────────────────────────────────────────────────

// ─────────────────────────────────────────────────────────────────────────────────

func MakeSubKeyAtter[T ItemInterface](obj T) (attr []string) {
	return append(
		[]string{
			obj.ItemKey().GetCollectionId(),
			obj.ItemKey().GetItemType(),
		},
		obj.ItemKey().GetItemKeyParts()...,
	)
}

func MakeSubItemKeyAtter(key *authpb.ItemKey) (attr []string) {
	return append(
		[]string{key.GetCollectionId(), key.GetItemType()},
		key.GetItemKeyParts()...,
	)
}

// ─────────────────────────────────────────────────────────────────────────────────

// func MakeHiddenKeyAtter[T common.ItemInterface](obj T) (attr []string) {
// 	return append([]string{common.HiddenItemType}, ...)
// }

func MakeHiddenKey[T ItemInterface](obj T) (hiddenKey string, err error) {
	return MakeComposeKey(
		HiddenItemType,
		MakeSubKeyAtter(obj),
	)
}

// ─────────────────────────────────────────────────────────────────────────────────

func MakeSuggestionKeyAtter[T ItemInterface](
	obj T,
	suggestionId string,
) (attr []string) {
	return append(MakeSubKeyAtter(obj), suggestionId)
}

// Key should be {<SUGGESTION>}{COLLECTION_ID}{ITEM_TYPE}{...ITEM_ID}{SuggestionId}
func MakeSuggestionPrimaryKey[T ItemInterface](
	obj T,
	suggestionId string,
) (suggestionKey string, err error) {
	return MakeComposeKey(
		SuggestionItemType,
		MakeSuggestionKeyAtter(obj, suggestionId),
	)
}

func MakeItemKeySuggestion(
	objKey *authpb.ItemKey,
	suggestionId string,
) (suggestionKey string) {
	subKey := KeyToSubKey(objKey, SuggestionItemType)
	subKey.ItemKeyParts = append(subKey.GetItemKeyParts(), suggestionId)

	return MakeStateKey(subKey)
}

func MakeItemKeySuggestionKeyAttr(
	objKey *authpb.ItemKey,
	suggestionId string,
) (attr []string) {
	return append(MakeSubItemKeyAtter(objKey), suggestionId)
}

// ─────────────────────────────────────────────────────────────────────────────────
func MakeRefKeyAttrs(
	ref *authpb.ReferenceKey,
) (refKey1 []string, refKey2 []string, err error) {
	if ref == nil || (ref.GetKey1() == nil && ref.GetKey2() == nil) {
		return refKey1, refKey2, oops.Errorf("Invalid reference")
	}

	var a, b []string

	if ref.GetKey1() != nil {
		// a = append([]string{ref.Key_1.GetCollectionId(), ref.GetKey_1().GetItemType()}, ref.GetKey_1().GetItemKeyParts()...)
		a = MakeSubItemKeyAtter(ref.GetKey1())
	}
	if ref.GetKey2() != nil {
		// b = append([]string{ref.GetKey_2().GetItemType()}, ref.GetKey_2().GetItemKeyParts()...)
		b = MakeSubItemKeyAtter(ref.GetKey2())
	}
	refKey1 = slices.Clone(a)
	refKey1 = append(refKey1, b...)

	refKey2 = slices.Clone(b)
	refKey2 = append(refKey2, a...)

	return refKey1, refKey2, nil
}

func MakeRefKeys(
	ref *authpb.ReferenceKey,
) (refKey1 string, refKey2 string, err error) {
	// attr := obj.KeyAttr()
	// ItemKey := obj.ItemKey()

	if ref == nil || (ref.GetKey1() == nil && ref.GetKey2() == nil) {
		return "", "", oops.Errorf("Invalid reference")
	}

	var a, b, k1, k2 []string

	if ref.GetKey1() != nil {
		// a = append([]string{ref.Key_1.GetCollectionId(), ref.GetKey_1().GetItemType()}, ref.GetKey_1().GetItemKeyParts()...)
		a = MakeSubItemKeyAtter(ref.GetKey1())
	}
	if ref.GetKey2() != nil {
		// b = append([]string{ref.GetKey_2().GetItemType()}, ref.GetKey_2().GetItemKeyParts()...)
		b = MakeSubItemKeyAtter(ref.GetKey2())
	}

	switch {
	case ref.GetKey1() != nil && ref.GetKey2() != nil:
		{
			k1 = slices.Clone(a)
			k1 = append(k1, b...)
			k2 = slices.Clone(b)
			k2 = append(k2, a...)

			refKey1, err = MakeComposeKey(ReferenceItemType, k1)
			if err != nil {
				return "", "", err
			}

			refKey2, err = MakeComposeKey(ReferenceItemType, k2)
			if err != nil {
				return "", "", err
			}

			return refKey1, refKey2, nil
		}
	case ref.GetKey1() != nil && ref.GetKey2() == nil:
		{
			refKey1, err = MakeComposeKey(ReferenceItemType, a)
			if err != nil {
				return "", "", err
			}
			return refKey1, "", nil

		}
	case ref.GetKey1() == nil && ref.GetKey2() != nil:
		{
			refKey2, err = MakeComposeKey(ReferenceItemType, b)
			if err != nil {
				return "", "", err
			}

			return "", refKey2, nil
		}
	default:
		return "", "", oops.Errorf("Invalid reference")
	}
}
