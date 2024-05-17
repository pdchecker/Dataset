package main

import (
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/nova38/saacs/pkg/chaincode/common"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

// UTIL Functions

// Returns true if the key exists in the ledger
func KeyExists(ctx common.TxCtxInterface, key string) bool {
	bytes, err := ctx.GetStub().GetState(key)
	if bytes == nil && err == nil {
		return false
	}

	return err == nil
}

func Exists[T common.ItemInterface](ctx common.TxCtxInterface, obj T) bool {
	return KeyExists(ctx, obj.StateKey())
}

// GetFromKey returns the item from the ledger
func GetFromKey[T common.ItemInterface](ctx common.TxCtxInterface, key string, obj T) (err error) {
	bytes, err := ctx.GetStub().GetState(key)

	if bytes == nil && err == nil {
		return oops.
			With("Key", key, "ItemType", obj.ItemType()).
			Wrap(common.KeyNotFound)
	} else if err != nil {
		return oops.
			With("Key", key, "ItemType", obj.ItemType()).
			Wrap(err)
	}

	if err = json.Unmarshal(bytes, obj); err != nil {
		return oops.With(
			"Key", obj.StateKey(),
			"ItemType", obj.ItemType(),
		).Wrap(err)
	}

	return nil
}

// Get returns the item from the ledger
func Get[T common.ItemInterface](ctx common.TxCtxInterface, obj T) (err error) {

	bytes, err := ctx.GetStub().GetState(obj.StateKey())

	if bytes == nil && err == nil {
		return oops.
			With(
				"Key", obj.StateKey(),
				"ItemType", obj.ItemType(),
				"ItemKey", obj.ItemKey(),
			).
			Wrap(common.KeyNotFound)
	} else if err != nil {
		return oops.
			With(
				"Key", obj.StateKey(),
				"ItemType", obj.ItemType(),
				"ItemKey", obj.ItemKey(),
			).
			Wrap(err)
	}

	if err = json.Unmarshal(bytes, obj); err != nil {
		return oops.With(
			"Key", obj.StateKey(),
			"ItemType", obj.ItemType(),
		).Wrap(err)
	}

	return nil
}

func Insert[T common.ItemInterface](ctx common.TxCtxInterface, obj T) (err error) {
	if exists := Exists(ctx, obj); exists {
		return oops.With(
			"Key", obj.StateKey(),
			"ItemType", obj.ItemType(),
		).Wrap(common.AlreadyExists)
	}

	if err := Put[T](ctx, obj); err != nil {
		return oops.With(
			"Key", obj.StateKey(),
			"ItemType", obj.ItemType(),
		).Wrap(err)
	}

	return nil
}

// Put puts the item into the ledger after json marshalling it
func Put[T common.ItemInterface](ctx common.TxCtxInterface, obj T) (err error) {

	if bytes, err := json.Marshal(obj); err != nil {
		return oops.Hint("Failed to Marshal").With(
			"Key", obj.StateKey(),
			"ItemType", obj.ItemType(),
		).Wrap(err)
	} else {
		err := ctx.GetStub().PutState(obj.StateKey(), bytes)
		if err != nil {
			return oops.With(
				"Key", obj.StateKey(),
				"ItemType", obj.ItemType(),
			).Wrap(err)
		}
	}
	return nil
}

func Delete[T common.ItemInterface](ctx common.TxCtxInterface, obj T) (err error) {
	key := obj.StateKey()

	if err = ctx.GetStub().DelState(key); err != nil {
		return oops.With(
			"Key", obj.StateKey(),
			"ItemType", obj.ItemType(),
		).Wrap(err)
	}

	return nil
}

// ════════════════════════════════════════════════════════
// Invoke Functions
// ════════════════════════════════════════════════════════

// Insert inserts the item into the ledger
// returns error if the item already exists

// ════════════════════════════════════════════════════════
// Query Functions
// ════════════════════════════════════════════════════════

// func (l *Ledger[T]) GetFromKey(key string) (obj T, err error) {
// 	return GetFromKey[T](l.ctx, key)
// }

// GetPartialKeyList returns a list of items of type T
// T must implement StateItem interface
// numAttr is the number of attributes in the key to search for
func GetPartialKeyList[T common.ItemInterface](
	ctx common.TxCtxInterface,
	obj T,
	numAttr int,
	bookmark string,
) (list []T, mk string, err error) {
	// obj = []*T{}
	ctx.GetLogger().Info("GetPartialKeyList")

	var (
		itemtype = obj.ItemType()
		attr     = obj.KeyAttr()
	)

	if len(attr) == 0 || len(attr) < numAttr {
		return nil, "", oops.Wrap(common.ItemInvalid)
	}

	// Extract the attributes to search for
	attr = lo.DropRight(attr, numAttr)

	ctx.GetLogger().
		Debug("GetPartialKeyList",
			slog.Group(
				"Key", "ItemType", itemtype,
				slog.Int("numAttr", numAttr),
				slog.Any("attr", attr),
				slog.Group(
					"Paged",
					"Bookmark", bookmark,
					"PageSize", strconv.Itoa(int(ctx.GetPageSize())),
				),
			),
		)

	results, meta, err := ctx.GetStub().
		GetStateByPartialCompositeKeyWithPagination(
			obj.ItemKey().GetItemType(),
			attr,
			ctx.GetPageSize(),
			bookmark,
		)
	if err != nil {
		return nil, "", err
	}
	defer func(results shim.StateQueryIteratorInterface) {
		err := results.Close()
		if err != nil {
			ctx.GetLogger().Error("GetPartialKeyList", "Error", err)
		}
	}(results)

	for results.HasNext() {

		tmpObj, ok := obj.ProtoReflect().New().Interface().(T)
		if !ok {
			return nil, "", oops.Errorf("Error cloning object")
		}

		queryResponse, err := results.Next()
		if err != nil || queryResponse == nil {
			return nil, "", oops.Wrapf(err, "Error getting next item")
		}

		if err = json.Unmarshal(queryResponse.GetValue(), tmpObj); err != nil {
			return nil, "", oops.Wrap(err)
		}

		list = append(list, tmpObj)
	}

	return list, meta.GetBookmark(), nil
}

// ════════════════════════════════════════════════════════
// Raw Functions
// ════════════════════════════════════════════════════════

// func GetFromKey[T common.ItemInterface](ctx common.TxCtxInterface, key string) (obj T, err error) {
// 	bytes, err := ctx.GetStub().GetState(key)
// 	if bytes == nil && err == nil {
// 		return obj, oops.
// 			With("Key", key, "ItemType", obj.ItemType()).
// 			Wrap(common.KeyNotFound)
// 	} else if err != nil {
// 		return obj, oops.Wrap(err)
// 	}

// 	if err = json.Unmarshal(bytes, obj); err != nil {
// 		return obj, oops.Wrap(err)
// 	}

// 	return obj, nil
// }

// bytes, err := l.ctx.GetStub().GetState(key)
// 	if bytes == nil && err == nil {
// 		return obj, oops.
// 			With("Key", key, "ItemType", obj.ItemType()).
// 			Wrap(common.KeyNotFound)
// 	} else if err != nil {
// 		return obj, oops.Wrap(err)
// 	}

// 	if err = json.Unmarshal(bytes, obj); err != nil {
// 		return obj, oops.Wrap(err)
// 	}

// 	return obj, nil

// ════════════════════════════════════════════════════════
// Invoke Functions
// ════════════════════════════════════════════════════════

// Insert inserts the item into the ledger
// returns error if the item already exists
// func Insert[T common.ItemInterface](ctx common.TxCtxInterface, obj T) (err error) {
//	var (
//		key   string
//		bytes []byte
//	)
//
//	if key, err = common.MakePrimaryKey(obj); err != nil {
//		return err
//	}
//
//	if Exists(ctx, key) {
//		return oops.
//			With("Key", key, "ItemType", obj.ItemType()).
//			Wrap(common.AlreadyExists)
//	}
//
//	if bytes, err = json.Marshal(obj); err != nil {
//		return err
//	}
//
//	return ctx.GetStub().PutState(key, bytes)
//}

// Edit updates the item in the ledger
// returns error if the item does not exist
// func Update[T common.ItemInterface](
//	ctx common.TxCtxInterface,
//	update T,
//	mask *fieldmaskpb.FieldMask,
// ) (obj T, err error) {
//	var (
//		key   string
//		bytes []byte
//	)
//
//}
//
//// Delete deletes the item from the ledger
// func Delete[T common.ItemInterface](ctx common.TxCtxInterface, in T) (err error) {
//	if err != nil {
//		return err
//	}
//
//	if err = ctx.GetStub().DelState(key); err != nil {
//		return oops.Wrap(err)
//	}
//
//	return nil
//}
