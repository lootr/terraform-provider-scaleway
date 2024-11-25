package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceOrderDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrderDomainCreate,
		ReadContext:   resourceOrderDomainsRead,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainRecordTimeout),
			Read:    schema.DefaultTimeout(defaultDomainRecordTimeout),
			Update:  schema.DefaultTimeout(defaultDomainRecordTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainRecordTimeout),
			Default: schema.DefaultTimeout(defaultDomainRecordTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain name to be managed",
			},
			"duration_in_years": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"project_id": account.ProjectIDSchema(),

			"owner_contact_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the owner contact. Either `owner_contact_id` or `owner_contact` must be provided.",
			},
			"owner_contact": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
				Description: "Details of the owner contact. Either `owner_contact_id` or `owner_contact` must be provided.",
			},
			"administrative_contact_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"administrative_contact": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
			},
			"technical_contact_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"technical_contact": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
			},
			//computed
			"auto_renew_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the automatic renewal of the domain.",
			},
			"dnssec_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the DNSSEC configuration of the domain.",
			},
			"epp_code": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of the domain's EPP codes.",
			},
			"expired_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of expiration of the domain (RFC 3339 format).",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last modification date of the domain (RFC 3339 format).",
			},
			"registrar": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The registrar managing the domain.",
			},
			"is_external": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether Scaleway is the domain's registrar.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the domain.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID associated with the domain.",
			},
			"pending_trade": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if a trade is ongoing for the domain.",
			},
			"external_domain_registration_status": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Registration status of an external domain, if applicable.",
			},
			"transfer_registration_status": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Status of the domain transfer, when available.",
			},
			"linked_products": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of Scaleway resources linked to the domain.",
			},
			"tld": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the TLD.",
						},
						"dnssec_support": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether DNSSEC is supported for this TLD.",
						},
						"duration_in_years_min": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Minimum duration (in years) for which this TLD can be registered.",
						},
						"duration_in_years_max": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum duration (in years) for which this TLD can be registered.",
						},
						"idn_support": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether this TLD supports IDN (Internationalized Domain Names).",
						},
						"offers": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of the offer action (e.g., create, transfer).",
									},
									"operation_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Path of the operation associated with the offer.",
									},
									"price": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"currency_code": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Currency code of the price.",
												},
												"units": {
													Type:        schema.TypeInt,
													Computed:    true,
													Description: "Unit price for the operation.",
												},
												"nanos": {
													Type:        schema.TypeInt,
													Computed:    true,
													Description: "Nano part of the price for more precision.",
												},
											},
										},
										Description: "Pricing information for the TLD offer.",
									},
								},
							},
							Description: "Available offers for the TLD.",
						},
						"specifications": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Specifications of the TLD.",
						},
					},
				},
				Description: "Details about the TLD (Top-Level Domain).",
			},
			"dns_zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeMap},
				Description: "List of DNS zones associated with the domain.",
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			hasOwnerContactID := d.HasChange("owner_contact_id") && d.Get("owner_contact_id").(string) != ""
			hasOwnerContact := d.HasChange("owner_contact") && len(d.Get("owner_contact").(map[string]interface{})) > 0

			if !hasOwnerContactID && !hasOwnerContact {
				return fmt.Errorf("either `owner_contact_id` or `owner_contact` must be provided")
			}

			if hasOwnerContactID && hasOwnerContact {
				return fmt.Errorf("only one of `owner_contact_id` or `owner_contact` can be provided")
			}

			return nil
		},
	}
}

func contactSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"legal_form": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Legal form of the contact (e.g., 'individual' or 'organization').",
		},
		"firstname": {
			Type:     schema.TypeString,
			Required: true,
		},
		"lastname": {
			Type:     schema.TypeString,
			Required: true,
		},
		"company_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"email": {
			Type:     schema.TypeString,
			Required: true,
		},
		"email_alt": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"phone_number": {
			Type:     schema.TypeString,
			Required: true,
		},
		"fax_number": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"address_line_1": {
			Type:     schema.TypeString,
			Required: true,
		},
		"address_line_2": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"zip": {
			Type:     schema.TypeString,
			Required: true,
		},
		"city": {
			Type:     schema.TypeString,
			Required: true,
		},
		"country": {
			Type:     schema.TypeString,
			Required: true,
		},
		"vat_identification_code": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"company_identification_code": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"lang": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"resale": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"extension_fr": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"extension_eu": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"whois_opt_in": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"state": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"extension_nl": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func resourceOrderDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	projectID := d.Get("project_id").(string)
	domainName := d.Get("domain_name").(string)
	durationInYears := uint32(d.Get("duration_in_years").(int))

	buyDomainsRequest := &domain.RegistrarAPIBuyDomainsRequest{
		Domains:         []string{domainName},
		DurationInYears: durationInYears,
		ProjectID:       projectID,
	}

	ownerContactID := d.Get("owner_contact_id").(string)
	if ownerContactID != "" {
		buyDomainsRequest.OwnerContactID = &ownerContactID
	} else if ownerContact, ok := d.GetOk("owner_contact"); ok {
		buyDomainsRequest.OwnerContact = ExpandNewContact(ownerContact.(map[string]interface{}))
	}

	adminContactID := d.Get("administrative_contact_id").(string)
	if adminContactID != "" {
		buyDomainsRequest.AdministrativeContactID = &adminContactID
	} else if adminContact, ok := d.GetOk("administrative_contact"); ok {
		buyDomainsRequest.AdministrativeContact = ExpandNewContact(adminContact.(map[string]interface{}))
	}

	techContactID := d.Get("technical_contact_id").(string)
	if techContactID != "" {
		buyDomainsRequest.TechnicalContactID = &techContactID
	} else if techContact, ok := d.GetOk("technical_contact"); ok {
		buyDomainsRequest.TechnicalContact = ExpandNewContact(techContact.(map[string]interface{}))
	}

	resp, err := registrarAPI.BuyDomains(buyDomainsRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ProjectID + "/" + domainName)

	return resourceOrderDomainsRead(ctx, d, m)
}

func resourceOrderDomainsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	id := d.Id()

	domainName, err := extractDomainFromID(id)
	if err != nil {
		return diag.FromErr(err)
	}

	getDomainRequest := &domain.RegistrarAPIGetDomainRequest{
		Domain: domainName,
	}

	res, err := registrarAPI.GetDomain(getDomainRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("domain_name", res.Domain); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("organization_id", res.OrganizationID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", res.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auto_renew_status", string(res.AutoRenewStatus)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expired_at", res.ExpiredAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", res.UpdatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("registrar", res.Registrar); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_external", res.IsExternal); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", string(res.Status)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pending_trade", res.PendingTrade); err != nil {
		return diag.FromErr(err)
	}

	// Mettre à jour les champs complexes
	if err := d.Set("owner_contact", flattenContact(res.OwnerContact)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("technical_contact", flattenContact(res.TechnicalContact)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("administrative_contact", flattenContact(res.AdministrativeContact)); err != nil {
		return diag.FromErr(err)
	}
	if res.Dnssec != nil {
		if err := d.Set("dnssec_status", string(res.Dnssec.Status)); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("epp_code", res.EppCode); err != nil {
		return diag.FromErr(err)
	}
	if res.Tld != nil {
		if err := d.Set("tld", flattenTLD(res.Tld)); err != nil {
			return diag.FromErr(err)
		}
	}
	if res.TransferRegistrationStatus != nil {
		if err := d.Set("transfer_registration_status", flattenDomainRegistrationStatusTransfer(res.TransferRegistrationStatus)); err != nil {
			return diag.FromErr(err)
		}
	}
	if res.ExternalDomainRegistrationStatus != nil {
		if err := d.Set("external_domain_registration_status", flattenExternalDomainRegistrationStatus(res.ExternalDomainRegistrationStatus)); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("linked_products", res.LinkedProducts); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dns_zones", flattenDNSZones(res.DNSZones)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
