# Create a reusable instance template for PyTorch training
resource "vastai_template" "pytorch_training" {
  name  = "pytorch-training-template"
  image = "pytorch/pytorch:2.1.0-cuda12.1-cudnn8-runtime"

  use_ssh         = true
  use_jupyter_lab = true
  ssh_direct      = true

  env = "{\"PYTHONUNBUFFERED\": \"1\", \"CUDA_VISIBLE_DEVICES\": \"all\"}"

  onstart_cmd     = "pip install wandb tensorboard && cd /workspace"
  readme          = "PyTorch training template with SSH and JupyterLab access."
  readme_visible  = true
  private         = false
}
