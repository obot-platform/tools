package graph

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ListAllContacts(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.Contactable, error) {
	contacts, err := client.Me().Contacts().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	return contacts.GetValue(), nil
}

func CreateContact(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, givenName, surname, email, businessPhone string) (models.Contactable, error) {
	requestBody := models.NewContact()
	requestBody.SetGivenName(&givenName) 
	requestBody.SetSurname(&surname) 

	emailAddress := models.NewEmailAddress()
	address := email
	emailAddress.SetAddress(&address) 
	name := givenName + " " + surname
	emailAddress.SetName(&name) 

	emailAddresses := []models.EmailAddressable {
		emailAddress,
	}
	requestBody.SetEmailAddresses(emailAddresses)
	businessPhones := []string {
		businessPhone,
	}
	requestBody.SetBusinessPhones(businessPhones)

	// To initialize your graphClient, see https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
	contacts, err := client.Me().Contacts().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact %s %s: %w", givenName, surname, err)
	}

	return contacts, nil
}

func GetContact(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, contactId string) (models.Contactable, error) {
	contact, err := client.Me().Contacts().ByContactId(contactId).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact by id %s: %w", contactId, err)
	}
	return contact, nil
}

func DeleteContact(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, contactId string) error {
	err := client.Me().Contacts().ByContactId(contactId).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete contact by id %s: %w", contactId, err)
	}
	return nil
}

func UpdateContact(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, contactId string, givenName, surname, email, businessPhone string) (models.Contactable, error) {
	requestBody := models.NewContact()
	requestBody.SetGivenName(&givenName) 
	requestBody.SetSurname(&surname) 
	emailAddress := models.NewEmailAddress()
	address := email
	emailAddress.SetAddress(&address) 
	name := givenName + " " + surname
	emailAddress.SetName(&name) 

	emailAddresses := []models.EmailAddressable {
		emailAddress,
	}
	requestBody.SetEmailAddresses(emailAddresses)
	businessPhones := []string {
		businessPhone,
	}
	requestBody.SetBusinessPhones(businessPhones)

	contacts, err := client.Me().Contacts().ByContactId(contactId).Patch(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact by id %s: %w", contactId, err)
	}
	return contacts, nil
}