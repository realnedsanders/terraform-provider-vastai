# Look up an existing instance by ID
data "vastai_instance" "example" {
  id = "123456"
}

output "instance_status" {
  value = data.vastai_instance.example.actual_status
}

output "instance_gpu" {
  value = data.vastai_instance.example.gpu_name
}
