output "workspace_name" {
  value = terraform.workspace
}

variable "accounts" {
  type = map(
    object({
      account_id    = string
      account_name  = string
      is_production = bool
    })
  )
}

locals {
  account_name = lower(replace(terraform.workspace, "_", "-"))
  account      = contains(keys(var.accounts), local.account_name) ? var.accounts[local.account_name]

  mandatory_moj_tags = {
    business-unit    = "OPG"
    application      = "opg-modernising-lpa"
    environment-name = local.account_name
    owner            = "OPG Webops: opgteam+modernising-lpa@digital.justice.gov.uk"
    is-production    = local.account.is_production
    runbook          = "https://github.com/ministryofjustice/opg-modernising-lpa"
    source-code      = "https://github.com/ministryofjustice/opg-modernising-lpa"
  }


  optional_tags = {
    infrastructure-support = "OPG Webops: opgteam+modernising-lpa@digital.justice.gov.uk"
  }

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)
}
