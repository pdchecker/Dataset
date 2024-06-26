package main

import (
	"fmt"
	"github.com/PM-Master/policy-machine-go/pdp"
	"github.com/PM-Master/policy-machine-go/policy"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/usnistgov/blossom/chaincode/adminmsp"
	"github.com/usnistgov/blossom/chaincode/collections"
	"github.com/usnistgov/blossom/chaincode/model"
	"github.com/usnistgov/blossom/chaincode/ngac/common"
	"github.com/usnistgov/blossom/chaincode/ngac/pap"
)

func InitCatalogNGAC(ctx contractapi.TransactionContextInterface) error {
	// check if this has already been called.  An error is thrown if this has not been called before
	if _, err := common.GetPvtCollPolicyStore(ctx, collections.Catalog()); err == nil {
		return fmt.Errorf("ngac initialization function has already been called")
	}

	mspid, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	// only member of the predefined AdminMSP can initialize the Catalog ngac store
	if mspid != adminmsp.AdminMSP {
		return fmt.Errorf("users in MSP %s do not have pemrission to initialize ngac graphs", mspid)
	}

	// check that the user is an admin user
	// only users with the attribute "blossom.admin=true" can call this function
	if err = ctx.GetClientIdentity().AssertAttributeValue(model.AdminAttribute, "true"); err != nil {
		return err
	}

	policyStore, err := pap.LoadCatalogPolicy()
	if err != nil {
		return fmt.Errorf("error loading catalog policy: %w", err)
	}

	return common.PutPvtCollPolicyStore(ctx, policyStore)
}

func CanRequestAccount(ctx contractapi.TransactionContextInterface) error {
	return ctx.GetClientIdentity().AssertAttributeValue(model.RoleAttribute, model.SystemOwnerRole)
}

func CanApproveAccount(ctx contractapi.TransactionContextInterface) error {
	return check(ctx, pap.BlossomObject, "approve_account")
}

func CanUploadATO(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "upload_ato")
}

func CanUpdateAccountStatus(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "update_account_status")
}

func CanRequestCheckout(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "check_out")
}

func CanApproveCheckout(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.BlossomObject, "approve_checkout")
}

func CanInitiateCheckIn(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "initiate_check_in")
}

func CanProcessCheckIn(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "process_check_in")
}

func CanReportSwID(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "report_swid")
}

func CanDeleteSwID(ctx contractapi.TransactionContextInterface, account string) error {
	return check(ctx, pap.AccountObjectName(account), "delete_swid")
}

func CanOnboardAsset(ctx contractapi.TransactionContextInterface) error {
	return check(ctx, "assets", "onboard_asset")
}

func CanOffboardAsset(ctx contractapi.TransactionContextInterface) error {
	return check(ctx, "assets", "offboard_asset")
}

func CanViewAssetPrivate(ctx contractapi.TransactionContextInterface) error {
	return check(ctx, "all_assets", "view_asset_private")
}

func CanViewAssetPublic(ctx contractapi.TransactionContextInterface) error {
	return check(ctx, "all_assets", "view_asset_public")
}

func CanViewAssets(ctx contractapi.TransactionContextInterface) error {
	return check(ctx, "all_assets", "view_assets")
}

func check(ctx contractapi.TransactionContextInterface, target, permission string) error {
	user, err := common.GetUsername(ctx)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	policyStore, err := common.GetPvtCollPolicyStore(ctx, collections.Catalog())
	if err != nil {
		return err
	}

	account, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	userAttributes := make([]string, 0)
	if account == adminmsp.AdminMSP {
		// adminmsp users need to have the admin role
		if err = ctx.GetClientIdentity().AssertAttributeValue(model.AdminAttribute, "true"); err != nil {
			return fmt.Errorf("adminmsp users need to have the attribute blossom.admin=true")
		}

		userAttributes = append(userAttributes, pap.AdminUA())
	} else {
		// if user is not in the adminmsp, get the role they have in their account
		role, err := getRole(ctx)
		if err != nil {
			return err
		}

		userAttributes = append(userAttributes, role)
	}

	// assign the user to the account and role
	if _, err = policyStore.Graph().CreateNode(user, policy.User, nil, pap.AccountUA(account), userAttributes...); err != nil {
		return fmt.Errorf("error assigning user %s to user attributes %s: %v", user, userAttributes, err)
	}

	decider := pdp.NewDecider(policyStore.Graph(), policyStore.Prohibitions())
	if ok, err := decider.HasPermissions(user, target, permission); err != nil {
		return fmt.Errorf("error checking if user %s can %s on %s: %w", user, permission, target, err)
	} else if !ok {
		return fmt.Errorf("user %s does not have permission %s on %s", user, permission, target)
	}

	return nil
}

func getRole(ctx contractapi.TransactionContextInterface) (role string, err error) {
	var ok bool
	if role, ok, err = ctx.GetClientIdentity().GetAttributeValue(model.RoleAttribute); err != nil {
		return "", err
	} else if !ok {
		return "", fmt.Errorf("user does not have a role")
	}

	if !model.IsValidRole(role) {
		return "", fmt.Errorf("unrecognized role: %s", role)
	}

	return role, err
}
