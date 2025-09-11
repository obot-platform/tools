// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/jsonschema-go/jsonschema"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", ":3000", "HTTP address to listen on for streamable HTTP server")

// StaticTokenCredential implements azcore.TokenCredential
type StaticTokenCredential struct {
	token string
}

func (s StaticTokenCredential) GetToken(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// ContactMCPServer wraps the Microsoft Graph client for Contact operations
type ContactMCPServer struct {
	client *msgraphsdkgo.GraphServiceClient
}

// NewContactMCPServer creates a new Contact MCP server with the given token
func NewContactMCPServer(token string) (*ContactMCPServer, error) {
	credential := StaticTokenCredential{token: token}
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &ContactMCPServer{client: client}, nil
}

// Argument structures with proper JSON schema tags based on tool.gpt
type ListContactsArgs struct{}

type CreateContactArgs struct {
	GivenName      *string `json:"given_name,omitempty" jsonschema:"(optional) The given name of the contact."`
	Surname        *string `json:"surname,omitempty" jsonschema:"(optional) The surname of the contact."`
	Emails         *string `json:"emails,omitempty" jsonschema:"(optional) a comma separated list of email addresses of the contact. Example: 'john.doe@example.com,jane.doe@example.com'"`
	BusinessPhones *string `json:"business_phones,omitempty" jsonschema:"(optional) a comma separated list of business phone numbers of the contact. Example: '+1234567890,+1234567891'"`
}

type UpdateContactArgs struct {
	ContactID      string  `json:"contact_id" jsonschema:"(required) The ID of the contact to update."`
	GivenName      *string `json:"given_name,omitempty" jsonschema:"(optional) Replace the given name of the contact."`
	Surname        *string `json:"surname,omitempty" jsonschema:"(optional) Replace the surname of the contact."`
	Emails         *string `json:"emails,omitempty" jsonschema:"(optional) a comma separated list of email addresses of the contact. Example: 'john.doe@example.com,jane.doe@example.com'. This will replace the existing email addresses."`
	BusinessPhones *string `json:"business_phones,omitempty" jsonschema:"(optional) a comma separated list of business phone numbers of the contact. Example: '+1234567890,+1234567891'. This will replace the existing business phone numbers."`
}

type DeleteContactArgs struct {
	ContactID string `json:"contact_id" jsonschema:"(required) The ID of the contact to delete."`
}

type GetContactArgs struct {
	ContactID string `json:"contact_id" jsonschema:"(required) The ID of the contact to get."`
}

// Contact information structure
type ContactInfo struct {
	ID             string   `json:"id"`
	GivenName      string   `json:"given_name,omitempty"`
	Surname        string   `json:"surname,omitempty"`
	DisplayName    string   `json:"display_name,omitempty"`
	EmailAddresses []string `json:"email_addresses,omitempty"`
	BusinessPhones []string `json:"business_phones,omitempty"`
}

// Helper functions
func deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}

func parseCommaSeparatedString(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ListContacts lists all Outlook contacts available to the user
func (c *ContactMCPServer) ListContacts(ctx context.Context, req *mcp.CallToolRequest, args ListContactsArgs) (*mcp.CallToolResult, any, error) {
	contacts, err := c.client.Me().Contacts().Get(ctx, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list contacts: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	var contactInfos []ContactInfo
	for _, contact := range contacts.GetValue() {
		contactInfo := ContactInfo{
			ID:          deref(contact.GetId()),
			DisplayName: deref(contact.GetDisplayName()),
			GivenName:   deref(contact.GetGivenName()),
			Surname:     deref(contact.GetSurname()),
		}

		// Extract email addresses
		if emails := contact.GetEmailAddresses(); len(emails) > 0 {
			for _, email := range emails {
				if email != nil && email.GetAddress() != nil {
					contactInfo.EmailAddresses = append(contactInfo.EmailAddresses, deref(email.GetAddress()))
				}
			}
		}

		// Extract business phones
		if phones := contact.GetBusinessPhones(); len(phones) > 0 {
			contactInfo.BusinessPhones = append(contactInfo.BusinessPhones, phones...)
		}

		contactInfos = append(contactInfos, contactInfo)
	}

	result, err := json.MarshalIndent(contactInfos, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal contact data: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// CreateContact creates a new Outlook contact
func (c *ContactMCPServer) CreateContact(ctx context.Context, req *mcp.CallToolRequest, args CreateContactArgs) (*mcp.CallToolResult, any, error) {
	// Parse emails and phone numbers
	var emails []string
	var phones []string

	if args.Emails != nil {
		emails = parseCommaSeparatedString(*args.Emails)
	}

	if args.BusinessPhones != nil {
		phones = parseCommaSeparatedString(*args.BusinessPhones)
	}

	// Validate that at least one field is provided
	if (args.GivenName == nil || *args.GivenName == "") && 
	   (args.Surname == nil || *args.Surname == "") && 
	   len(emails) == 0 && len(phones) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No parameters provided. Please provide at least one of the following parameters: given_name, surname, emails, business_phones",
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Build the contact request
	requestBody := models.NewContact()
	if args.GivenName != nil && *args.GivenName != "" {
		requestBody.SetGivenName(args.GivenName)
	}
	if args.Surname != nil && *args.Surname != "" {
		requestBody.SetSurname(args.Surname)
	}

	// Set email addresses
	if len(emails) > 0 {
		emailAddresses := []models.EmailAddressable{}
		for _, email := range emails {
			emailAddress := models.NewEmailAddress()
			address := email
			emailAddress.SetAddress(&address)
			emailAddresses = append(emailAddresses, emailAddress)
		}
		requestBody.SetEmailAddresses(emailAddresses)
	}

	// Set business phones
	if len(phones) > 0 {
		requestBody.SetBusinessPhones(phones)
	}

	// Create the contact
	contact, err := c.client.Me().Contacts().Post(ctx, requestBody, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to create contact: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Contact created successfully. ID: %s", deref(contact.GetId())),
			},
		},
	}, nil, nil
}

// UpdateContact updates an existing Outlook contact
func (c *ContactMCPServer) UpdateContact(ctx context.Context, req *mcp.CallToolRequest, args UpdateContactArgs) (*mcp.CallToolResult, any, error) {
	// Parse emails and phone numbers
	var emails []string
	var phones []string

	if args.Emails != nil {
		emails = parseCommaSeparatedString(*args.Emails)
	}

	if args.BusinessPhones != nil {
		phones = parseCommaSeparatedString(*args.BusinessPhones)
	}

	// Build the update request
	requestBody := models.NewContact()
	if args.GivenName != nil && *args.GivenName != "" {
		requestBody.SetGivenName(args.GivenName)
	}
	if args.Surname != nil && *args.Surname != "" {
		requestBody.SetSurname(args.Surname)
	}

	// Set email addresses if provided
	if len(emails) > 0 {
		emailAddresses := []models.EmailAddressable{}
		for _, email := range emails {
			emailAddress := models.NewEmailAddress()
			address := email
			emailAddress.SetAddress(&address)
			emailAddresses = append(emailAddresses, emailAddress)
		}
		requestBody.SetEmailAddresses(emailAddresses)
	}

	// Set business phones if provided
	if len(phones) > 0 {
		requestBody.SetBusinessPhones(phones)
	}

	// Update the contact
	contact, err := c.client.Me().Contacts().ByContactId(args.ContactID).Patch(ctx, requestBody, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to update contact by id %s: %v", args.ContactID, err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Contact updated successfully. ID: %s", deref(contact.GetId())),
			},
		},
	}, nil, nil
}

// DeleteContact deletes an existing Outlook contact
func (c *ContactMCPServer) DeleteContact(ctx context.Context, req *mcp.CallToolRequest, args DeleteContactArgs) (*mcp.CallToolResult, any, error) {
	err := c.client.Me().Contacts().ByContactId(args.ContactID).Delete(ctx, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to delete contact by id %s: %v", args.ContactID, err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Contact %s deleted successfully", args.ContactID),
			},
		},
	}, nil, nil
}

// GetContact gets the details of an existing Outlook contact
func (c *ContactMCPServer) GetContact(ctx context.Context, req *mcp.CallToolRequest, args GetContactArgs) (*mcp.CallToolResult, any, error) {
	contact, err := c.client.Me().Contacts().ByContactId(args.ContactID).Get(ctx, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get contact by id %s: %v", args.ContactID, err),
				},
			},
			IsError: true,
		}, nil, err
	}

	contactInfo := ContactInfo{
		ID:          deref(contact.GetId()),
		DisplayName: deref(contact.GetDisplayName()),
		GivenName:   deref(contact.GetGivenName()),
		Surname:     deref(contact.GetSurname()),
	}

	// Extract email addresses
	if emails := contact.GetEmailAddresses(); len(emails) > 0 {
		for _, email := range emails {
			if email != nil && email.GetAddress() != nil {
				contactInfo.EmailAddresses = append(contactInfo.EmailAddresses, deref(email.GetAddress()))
			}
		}
	}

	// Extract business phones
	if phones := contact.GetBusinessPhones(); len(phones) > 0 {
		contactInfo.BusinessPhones = append(contactInfo.BusinessPhones, phones...)
	}

	result, err := json.MarshalIndent(contactInfo, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal contact data: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// ExtractTokenFromRequest extracts the bearer token from HTTP request headers
func ExtractTokenFromRequest(req *http.Request) (string, error) {
	// Try X-Forwarded-Access-Token first
	if token := req.Header.Get("X-Forwarded-Access-Token"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no access token found in request headers")
}

func main() {
	flag.Parse()

	// Create server factory that extracts token from each request
	serverFactory := func(req *http.Request) *mcp.Server {
		token, err := ExtractTokenFromRequest(req)
		if err != nil {
			log.Printf("Failed to extract token from request: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "contact-mcp-server"}, nil)
			return server
		}

		contactServer, err := NewContactMCPServer(token)
		if err != nil {
			log.Printf("Failed to create Contact MCP server: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "contact-mcp-server"}, nil)
			return server
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "contact-mcp-server"}, nil)

		// Create JSON schemas for the tools
		createContactSchema, _ := jsonschema.For[CreateContactArgs](nil)
		updateContactSchema, _ := jsonschema.For[UpdateContactArgs](nil)
		deleteContactSchema, _ := jsonschema.For[DeleteContactArgs](nil)
		getContactSchema, _ := jsonschema.For[GetContactArgs](nil)

		// Register all tools with proper schemas - matching tool.gpt exactly
		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_contacts",
			Description: "List all Outlook contacts available to the user.",
		}, contactServer.ListContacts)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "create_contact",
			Description: "Create a new Outlook contact.",
			InputSchema: createContactSchema,
		}, contactServer.CreateContact)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "update_contact",
			Description: "Update an existing Outlook contact.",
			InputSchema: updateContactSchema,
		}, contactServer.UpdateContact)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "delete_contact",
			Description: "Delete an existing Outlook contact.",
			InputSchema: deleteContactSchema,
		}, contactServer.DeleteContact)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_contact",
			Description: "Get the details of an existing Outlook contact.",
			InputSchema: getContactSchema,
		}, contactServer.GetContact)

		return server
	}

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(serverFactory, nil)
		log.Printf("Contact MCP server listening at %s", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("HTTP address is required")
	}
}
