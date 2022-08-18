# route53 a record for *.noirgate.${domain}
resource "aws_route53_record" "wildcard_noirgate_subdomain" {
  zone_id = var.noirgate_zone_id
  #   zone_id = aws_route53_zone.noirgate.zone_id
  name    = "*.noirgate.${var.noirgate_domain}"
  type    = "A"
  ttl     = "60"
  records = [aws_instance.noirgate-host.public_ip]
}

resource "aws_route53_record" "noirgate_subdomain" {
  zone_id = var.noirgate_zone_id
  name    = "noirgate.${var.noirgate_domain}"
  type    = "A"
  ttl     = "60"
  records = [aws_instance.noirgate-host.public_ip]
}

# # acme txt record 
resource "aws_route53_record" "acme_txt_record" {
  zone_id = var.noirgate_zone_id
  name    = "_acme-challenge.noirgate.${var.noirgate_domain}"
  type    = "TXT"
  ttl     = "60"
  records = "PLACE_HOLDER"
  lifecycle {
    ignore_changes = ["records"]
  }
}
