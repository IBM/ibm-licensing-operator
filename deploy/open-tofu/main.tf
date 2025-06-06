module "ibm_license_service" {
  source           = "./component"
  chart_name       = "ibm-license-service"
  component_dirname = "license-service"
  namespace = "ibm-licensing"
}

module "ibm_license_service_reporter" {
  source           = "./component"
  chart_name       = "ibm-license-service-reporter"
  component_dirname = "reporter"
  namespace = "ibm-ls-reporter"
}

module "ibm_license_service_scanner" {
  source           = "./component"
  chart_name       = "ibm-license-service-scanner"
  component_dirname = "scanner"
  namespace = "ibm-licensing-scanner"
}
