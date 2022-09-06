data "aws_backup_vault" "main" {
  name     = "${local.environment.account_name}_main_backup_vault"
  provider = aws.eu_west_1
}

data "aws_iam_role" "aws_backup_role" {
  name     = "aws_backup_role"
  provider = aws.eu_west_1
}
