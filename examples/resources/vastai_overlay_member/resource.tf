# Join an instance to an overlay network
# NOTE: Destroying this resource removes it from Terraform state only;
# individual instance removal from overlays is not supported by the API.
resource "vastai_overlay_member" "training_instance" {
  overlay_name = vastai_overlay.training_net.name
  instance_id  = vastai_instance.training.id
}
