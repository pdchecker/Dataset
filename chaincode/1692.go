package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-samples/chaincode/thinh-chaincode/utils"
)

// ProductContract contract for handling products
type ProductContract struct {
	contractapi.Contract
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Product  `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

const index = "category~name"
const docType = "product"

// CreateProduct creates a new product and adds it to the world state using id as key.
// Note: A product created by this method is marked as unavailable upon creation time.
// In order to make it available, method MarkAsAvailable() must be used.
func (s *ProductContract) CreateProduct(ctx utils.RoleBasedTransactionContextInterface, id string, name string, origin string, category string,
	unitPrice float64, unitMeasurement string, quantity int, productionDateString string, expirationDateString string, imageSrc string) error {

	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to create products")
	}

	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return errors.New("unable to interact with world state")
	} else if existing != nil {
		return fmt.Errorf("cannot create new product in world state as id %s already exists", id)
	}

	product := Product{
		DocType:         docType,
		ID:              id,
		Name:            name,
		Category:        category,
		Origin:          origin,
		UnitPrice:       unitPrice,
		UnitMeasurement: unitMeasurement,
		Quantity:        quantity,
		ProductionDate:  utils.ParseTimeDefault(productionDateString),
		ExpirationDate:  utils.ParseTimeDefault(expirationDateString),
		ImageSource:     imageSrc,
		Available:       false,
	}

	productBytes, _ := json.Marshal(product)

	err = ctx.GetStub().PutState(id, []byte(productBytes))

	if err != nil {
		return errors.New("unable to interact with world state")
	}

	//  Create an index to enable category-based range queries.
	//  This will enable very efficient state range queries based on composite keys matching indexName~category~*
	categoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{product.Category, product.ID})
	if err != nil {
		return errors.New("error: cannot create index for category")
	}

	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the product.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	return ctx.GetStub().PutState(categoryNameIndexKey, value)
}

// CreateProductAvailable is the combination of CreateProduct() and MarkAsAvailable()
// This method creates a new product and marks it as available upon creation time.
// The product is then added to the world state using id as key.
func (s *ProductContract) CreateProductAvailable(ctx utils.RoleBasedTransactionContextInterface, id string, name string, origin string, category string,
	unitPrice float64, unitMeasurement string, quantity int, productionDateString string, expirationDateString string, imageSrc string) error {

	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to create products")
	}

	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return errors.New("unable to interact with world state")
	} else if existing != nil {
		return fmt.Errorf("cannot create new product in world state as id %s already exists", id)
	}

	product := Product{
		DocType:         docType,
		ID:              id,
		Name:            name,
		Category:        category,
		Origin:          origin,
		UnitPrice:       unitPrice,
		UnitMeasurement: unitMeasurement,
		Quantity:        quantity,
		ProductionDate:  utils.ParseTimeDefault(productionDateString),
		ExpirationDate:  utils.ParseTimeDefault(expirationDateString),
		ImageSource:     imageSrc,
		Available:       true,
	}

	productBytes, _ := json.Marshal(product)

	err = ctx.GetStub().PutState(id, []byte(productBytes))

	if err != nil {
		return errors.New("unable to interact with world state")
	}

	//  Create an index to enable category-based range queries.
	//  This will enable very efficient state range queries based on composite keys matching indexName~category~*
	categoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{product.Category, product.ID})
	if err != nil {
		return errors.New("error: cannot create index for category")
	}

	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the product.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	return ctx.GetStub().PutState(categoryNameIndexKey, value)
}

// ReadProduct retrieves a product from the ledger.
func (s *ProductContract) ReadProduct(ctx utils.RoleBasedTransactionContextInterface, productID string) (*Product, error) {
	productBytes, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product %s: %v", productID, err)
	}
	if productBytes == nil {
		return nil, fmt.Errorf("product %s does not exist", productID)
	}

	role := ctx.GetRole()

	var product Product
	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return nil, err
	} else if role != "Manufacturer" && !product.Available {
		return nil, fmt.Errorf("only manufacturers are allowed to read this product")
	}

	return &product, nil
}

// SaveUpdate saves a product to the ledger.
func SaveUpdate(ctx utils.RoleBasedTransactionContextInterface, product *Product) error {
	productID := product.ID
	productBytes, _ := json.Marshal(product)

	err := ctx.GetStub().PutState(productID, []byte(productBytes))

	if err != nil {
		return fmt.Errorf("failed to update product %s: %v", productID, err)
	}

	return nil
}

// UpdateProduct updates all product's fields using id as key.
// New value of a field may be the same as old value.
// However, if all new values are the same as old values, then no update is executed and nothing is committed to the ledger.
func (s *ProductContract) UpdateProduct(ctx utils.RoleBasedTransactionContextInterface, productID string,
	newName string, newOrigin string, newCategory string, newUnitPrice float64, newUnitMeasurement string,
	newQuantity int, newProductionDateString string, newExpirationDateString string) error {

	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	newProductionDate, newExpirationDate := utils.ParseTimeDefault(newProductionDateString), utils.ParseTimeDefault(newExpirationDateString)

	// flag to indicate whether ledger update and index update are needed
	isChanged, isIndexChanged := false, false

	// This variable is used only for storing the old value of the product category in case an index update is needed
	var oldCategory string

	/* Sequentially check for change in each field of a product */
	if product.Name != newName {
		product.Name = newName
		isChanged = true
	}
	if product.Origin != newOrigin {
		product.Origin = newOrigin
		isChanged = true
	}
	if product.Category != newCategory {
		oldCategory = product.Category
		product.Category = newCategory
		isChanged = true
		isIndexChanged = true
	}
	if product.UnitPrice != newUnitPrice {
		product.UnitPrice = newUnitPrice
		isChanged = true
	}
	if product.UnitMeasurement != newUnitMeasurement {
		product.UnitMeasurement = newUnitMeasurement
		isChanged = true
	}
	if product.Quantity != newQuantity {
		product.Quantity = newQuantity
		isChanged = true
	}
	if !product.ProductionDate.Equal(newProductionDate) {
		product.ProductionDate = newProductionDate
		isChanged = true
	}
	if !product.ExpirationDate.Equal(newExpirationDate) {
		product.ExpirationDate = newExpirationDate
		isChanged = true
	}

	// If there is no change, just return without any errors
	if !isChanged {
		return nil
	}

	// If there is no change in index, just save the update into the ledger
	if !isIndexChanged {
		return SaveUpdate(ctx, product)
	} else {
		/* If there is a change in index, also update the index */

		// First, save the update into the ledger
		err = SaveUpdate(ctx, product)
		if err != nil {
			return err
		}

		/* Then, update the index */

		// get old index entry key
		oldCategoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{oldCategory, product.ID})
		if err != nil {
			return err
		}

		// get new index entry key
		newCategoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{newCategory, product.ID})
		if err != nil {
			return err
		}

		// Delete old index entry
		err = ctx.GetStub().DelState(oldCategoryNameIndexKey)
		if err != nil {
			return err
		}

		// Create new index entry
		return ctx.GetStub().PutState(newCategoryNameIndexKey, []byte{0x00})
	}
}

// UpdateProductName updates a product's name using id as key.
func (s *ProductContract) UpdateProductName(ctx utils.RoleBasedTransactionContextInterface, productID string, newName string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.Name == newName {
		return nil
	}

	product.Name = newName

	return SaveUpdate(ctx, product)
}

// UpdateProductName updates a product's category using id as key.
func (s *ProductContract) UpdateProductCategory(ctx utils.RoleBasedTransactionContextInterface, productID string, newCategory string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.Category == newCategory {
		return nil
	}

	oldCategory := product.Category
	product.Category = newCategory

	err = SaveUpdate(ctx, product)
	if err != nil {
		return err
	}

	// get old index entry key
	oldCategoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{oldCategory, product.ID})
	if err != nil {
		return err
	}

	// get new index entry key
	newCategoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{newCategory, product.ID})
	if err != nil {
		return err
	}

	// Delete old index entry
	err = ctx.GetStub().DelState(oldCategoryNameIndexKey)
	if err != nil {
		return err
	}

	// Create new index entry
	return ctx.GetStub().PutState(newCategoryNameIndexKey, []byte{0x00})
}

// UpdateProductOrigin updates a product's origin using id as key.
func (s *ProductContract) UpdateProductOrigin(ctx utils.RoleBasedTransactionContextInterface, productID string, newOrigin string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.Origin == newOrigin {
		return nil
	}

	product.Origin = newOrigin

	return SaveUpdate(ctx, product)
}

// UpdateProductUnitPrice updates a product's unit price using id as key.
func (s *ProductContract) UpdateProductUnitPrice(ctx utils.RoleBasedTransactionContextInterface, productID string, newUnitPrice float64) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.UnitPrice == newUnitPrice {
		return nil
	}

	product.UnitPrice = newUnitPrice

	return SaveUpdate(ctx, product)
}

// UpdateProductUnitMeasurement updates a product's unit of measurement using id as key.
func (s *ProductContract) UpdateProductUnitMeasurement(ctx utils.RoleBasedTransactionContextInterface, productID string,
	newUnitMeasurement string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.UnitMeasurement == newUnitMeasurement {
		return nil
	}

	product.UnitMeasurement = newUnitMeasurement

	return SaveUpdate(ctx, product)
}

// UpdateProductQuantity updates a product's quantity using id as key.
func (s *ProductContract) UpdateProductQuantity(ctx utils.RoleBasedTransactionContextInterface, productID string, newQuantity int) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.Quantity == newQuantity {
		return nil
	}

	product.Quantity = newQuantity

	return SaveUpdate(ctx, product)
}

// UpdateProductProductionDate updates a product's production date using id as key.
func (s *ProductContract) UpdateProductProductionDate(ctx utils.RoleBasedTransactionContextInterface, productID string,
	newProductionDateString string) error {

	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	newProductionDate := utils.ParseTimeDefault(newProductionDateString)

	// If there is no change, just return without any errors
	if product.ProductionDate.Equal(newProductionDate) {
		return nil
	}

	product.ProductionDate = newProductionDate

	return SaveUpdate(ctx, product)
}

// UpdateProductExpirationDate updates a product's expiration date using id as key.
func (s *ProductContract) UpdateProductExpirationDate(ctx utils.RoleBasedTransactionContextInterface, productID string,
	newExpirationDateString string) error {

	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	newExpirationDate := utils.ParseTimeDefault(newExpirationDateString)

	// If there is no change, just return without any errors
	if product.ExpirationDate.Equal(newExpirationDate) {
		return nil
	}

	product.ExpirationDate = newExpirationDate

	return SaveUpdate(ctx, product)
}

// UpdateProductUnitMeasurement updates a product's image source using id as key.
func (s *ProductContract) UpdateProductImageSource(ctx utils.RoleBasedTransactionContextInterface, productID string, newImageSrc string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If there is no change, just return without any errors
	if product.ImageSource == newImageSrc {
		return nil
	}

	product.ImageSource = newImageSrc

	return SaveUpdate(ctx, product)
}

// MarkAsAvailable marks a product as available.
func (s *ProductContract) MarkAsAvailable(ctx utils.RoleBasedTransactionContextInterface, productID string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If the product is already available, just return without any errors
	if product.Available {
		return nil
	}

	product.Available = true

	return SaveUpdate(ctx, product)
}

// MarkAsAvailable marks a product as unavailable.
func (s *ProductContract) MarkAsUnavailable(ctx utils.RoleBasedTransactionContextInterface, productID string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to update products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	// If the product is already unavailable, just return without any errors
	if !product.Available {
		return nil
	}

	product.Available = false

	return SaveUpdate(ctx, product)
}

// DeleteProduct removes a product key-value pair from the ledger.
func (s *ProductContract) DeleteProduct(ctx utils.RoleBasedTransactionContextInterface, productID string) error {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return fmt.Errorf("only manufacturers are allowed to delete products")
	}

	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(productID)
	if err != nil {
		return fmt.Errorf("failed to delete product %s: %v", productID, err)
	}

	categoryNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{product.Category, product.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(categoryNameIndexKey)
}

// ProductExists returns true when product with given ID exists in the ledger.
func (s *ProductContract) ProductExists(ctx utils.RoleBasedTransactionContextInterface, productID string) (bool, error) {
	productBytes, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return false, fmt.Errorf("failed to read product %s from world state. %v", productID, err)
	}

	role := ctx.GetRole()
	// For manufacturers, they can see all products so only checking for productBytes is enough to verify that a product exists
	if role == "Manufacturer" {
		return productBytes != nil, nil

		// Otherwise,
	} else {
		// If the product does not exist on the ledger, productBytes will be nil; in that case, return false and no error
		if productBytes == nil {
			return false, nil
		}

		// For retailers, They can only see available products so additional check for availability of a product is necessary
		var product Product
		err = json.Unmarshal(productBytes, &product)
		if err != nil {
			return false, err
		}

		// If the product is not available, the product also does not exist from the perspective of the retailer
		if !product.Available {
			return false, nil

			// Otherwise, the product does indeed exist (from the perspective of the retailer)
		} else {
			return true, nil
		}
	}
}

// constructQueryResponseFromIterator constructs a slice of products from the resultsIterator.
// This function also takes the client's role into account when constructing the result
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface, role string) ([]*Product, error) {
	var products []*Product
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var product Product
		err = json.Unmarshal(queryResult.Value, &product)
		if err != nil {
			return nil, err
		}

		// Manufacturers can normally see all the products
		// For other roles, the product must be available in order to be seen
		if role == "Manufacturer" || product.Available {
			products = append(products, &product)
		}
	}

	return products, nil
}

// GetProductsByRange performs a range query based on the start and end IDs provided.
func (s *ProductContract) GetProductsByRange(ctx utils.RoleBasedTransactionContextInterface, startID, endID string) ([]*Product, error) {
	role := ctx.GetRole()

	resultsIterator, err := ctx.GetStub().GetStateByRange(startID, endID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator, role)
}

// GetProductsByCategory performs a range query based on the category provided.
func (s *ProductContract) GetProductsByCategory(ctx utils.RoleBasedTransactionContextInterface, category string) ([]*Product, error) {
	role := ctx.GetRole()

	categorizedProductResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{category})
	if err != nil {
		return nil, err
	}
	defer categorizedProductResultsIterator.Close()

	var products []*Product
	for categorizedProductResultsIterator.HasNext() {
		responseRange, err := categorizedProductResultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			returnedProductID := compositeKeyParts[1]
			product, err := s.ReadProduct(ctx, returnedProductID)
			if err != nil {
				return nil, err
			}

			// Manufacturers can normally see all the products
			// For other roles, the product must be available in order to be seen
			if role == "Manufacturer" || product.Available {
				products = append(products, product)
			}
		}
	}

	return products, nil
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx utils.RoleBasedTransactionContextInterface, queryString string) ([]*Product, error) {
	role := ctx.GetRole()

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator, role)
}

// QueryAllProducts queries for all products.
// Only available on state databases that support rich query (e.g. CouchDB)
func (s *ProductContract) QueryAllProducts(ctx utils.RoleBasedTransactionContextInterface) ([]*Product, error) {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return s.QueryAllAvailableProducts(ctx)
	}

	queryString := fmt.Sprintf(`{
									"selector":{
										"docType":"%s"
									}, 
									"use_index":[
										"_design/indexProductDoc", 
										"indexProduct"
									]
								}`,
		docType)
	return getQueryResultForQueryString(ctx, queryString)
}

// QueryAllAvailableProducts queries for all available products.
// Only available on state databases that support rich query (e.g. CouchDB)
func (s *ProductContract) QueryAllAvailableProducts(ctx utils.RoleBasedTransactionContextInterface) ([]*Product, error) {
	queryString := fmt.Sprintf(`{
									"selector":{
										"docType":"%s",
										"available":%t
									},
									"use_index":[
										"_design/indexProductAvailableDoc", 
										"indexProductAvailable"
									]
								}`,
		docType, true)
	return getQueryResultForQueryString(ctx, queryString)
}

// QueryAllUnavailableProducts queries for all unavailable products.
// Only available on state databases that support rich query (e.g. CouchDB)
func (s *ProductContract) QueryAllUnavailableProducts(ctx utils.RoleBasedTransactionContextInterface) ([]*Product, error) {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return nil, fmt.Errorf("only manufacturers are allowed to see unavailable products")
	}

	queryString := fmt.Sprintf(`{
									"selector":{
										"docType":"%s",
										"available":%t
									},
									"use_index":[
										"_design/indexProductAvailableDoc", 
										"indexProductAvailable"
									]
								}`,
		docType, false)
	return getQueryResultForQueryString(ctx, queryString)
}

// QueryProductsByOwner queries for products based on their names.
// Only available on state databases that support rich query (e.g. CouchDB)
func (s *ProductContract) QueryProductsByName(ctx utils.RoleBasedTransactionContextInterface, name string) ([]*Product, error) {
	queryString := fmt.Sprintf(`{
									"selector":{
										"docType":"%s",
										"name":"%s"
									}, 
									"use_index":[
										"_design/indexProductNameDoc", 
										"indexProductName"
									]
								}`,
		docType, name)
	return getQueryResultForQueryString(ctx, queryString)
}

// QueryProductsByCategory an alias of GetProductsByCategory()
func (s *ProductContract) QueryProductsByCategory(ctx utils.RoleBasedTransactionContextInterface, category string) ([]*Product, error) {
	return s.GetProductsByCategory(ctx, category)
}

// QueryProducts uses a query string to perform a query for products.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// Only available on state databases that support rich query (e.g. CouchDB)
func (s *ProductContract) QueryProducts(ctx utils.RoleBasedTransactionContextInterface, queryString string) ([]*Product, error) {
	return getQueryResultForQueryString(ctx, queryString)
}

// GetProductHistory returns the chain of update for a product since issuance.
func (s *ProductContract) GetProductHistory(ctx utils.RoleBasedTransactionContextInterface, productID string) ([]HistoryQueryResult, error) {
	role := ctx.GetRole()
	if role != "Manufacturer" {
		return nil, fmt.Errorf("only manufacturers are allowed to see products log")
	}

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(productID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &product)
			if err != nil {
				return nil, err
			}
		} else {
			product = Product{
				ID: productID,
			}
		}

		timestamp, err := response.Timestamp.AsTime(), response.Timestamp.CheckValid()
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &product,
			IsDelete:  response.IsDelete,
		}

		records = append(records, record)
	}

	return records, nil
}

// GetEvaluateTransactions returns functions of ProductContract not to be tagged as submit.
func (s *ProductContract) GetEvaluateTransactions() []string {
	return []string{"ReadProduct", "ProductExists", "GetProductsByRange", "GetProductsByCategory", "QueryAllProducts", "QueryAllAvailableProducts",
		"QueryAllUnavailableProducts", "QueryProductsByName", "QueryProductsByCategory", "QueryProducts", "GetProductHistory"}
}
