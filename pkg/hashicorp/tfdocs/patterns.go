// pkg/hashicorp/tfdocs/patterns.go
package tfdocs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PatternCategory represents a category of Terraform patterns
type PatternCategory string

const (
	CategoryCompute     PatternCategory = "compute"
	CategoryNetworking  PatternCategory = "networking"
	CategoryStorage     PatternCategory = "storage"
	CategoryDatabase    PatternCategory = "database"
	CategorySecurity    PatternCategory = "security"
	CategoryApplication PatternCategory = "application"
	CategoryMonitoring  PatternCategory = "monitoring"
)

// CloudProvider represents a cloud provider
type CloudProvider string

const (
	ProviderAWS    CloudProvider = "aws"
	ProviderAzure  CloudProvider = "azure"
	ProviderGCP    CloudProvider = "gcp"
	ProviderGeneric CloudProvider = "generic"
)

// ComplexityLevel represents the complexity level of a pattern
type ComplexityLevel string

const (
	ComplexityBasic       ComplexityLevel = "basic"
	ComplexityIntermediate ComplexityLevel = "intermediate"
	ComplexityAdvanced    ComplexityLevel = "advanced"
)

// Pattern represents a Terraform code pattern
type Pattern struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    PatternCategory  `json:"category"`
	Provider    CloudProvider    `json:"provider"`
	Complexity  ComplexityLevel  `json:"complexity"`
	Files       map[string]string `json:"files"`
	Tags        []string         `json:"tags"`
}

// PatternFilter defines filtering criteria for patterns
type PatternFilter struct {
	Category   *PatternCategory
	Provider   *CloudProvider
	Complexity *ComplexityLevel
	Tags       []string
	Query      string
}

// PatternRepository manages Terraform pattern templates
type PatternRepository struct {
	patterns      map[string]*Pattern
	patternPath   string
	mutex         sync.RWMutex
	logger        Logger
}

// NewPatternRepository creates a new pattern repository
func NewPatternRepository(patternPath string, logger Logger) *PatternRepository {
	return &PatternRepository{
		patterns:    make(map[string]*Pattern),
		patternPath: patternPath,
		logger:      logger,
	}
}

// Initialize loads patterns from the pattern directory
func (r *PatternRepository) Initialize() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.logger.Info("Initializing pattern repository", "path", r.patternPath)

	// Create pattern directory if it doesn't exist
	if err := os.MkdirAll(r.patternPath, 0755); err != nil {
		return fmt.Errorf("failed to create pattern directory: %w", err)
	}

	// Read pattern index file
	indexPath := filepath.Join(r.patternPath, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		r.logger.Info("Pattern index file not found, initializing with default patterns")
		return r.initializeDefaultPatterns()
	}

	// Read pattern index
	data, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read pattern index: %w", err)
	}

	var patterns []*Pattern
	if err := json.Unmarshal(data, &patterns); err != nil {
		return fmt.Errorf("failed to parse pattern index: %w", err)
	}

	// Load patterns
	for _, pattern := range patterns {
		patternDir := filepath.Join(r.patternPath, pattern.ID)
		if err := r.loadPattern(pattern, patternDir); err != nil {
			r.logger.Error("Failed to load pattern", "id", pattern.ID, "error", err)
			continue
		}
		r.patterns[pattern.ID] = pattern
	}

	r.logger.Info("Pattern repository initialized", "count", len(r.patterns))
	return nil
}

// loadPattern loads a pattern's files from disk
func (r *PatternRepository) loadPattern(pattern *Pattern, patternDir string) error {
	// Read pattern files
	files, err := ioutil.ReadDir(patternDir)
	if err != nil {
		return fmt.Errorf("failed to read pattern directory: %w", err)
	}

	pattern.Files = make(map[string]string)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(patternDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read pattern file: %w", err)
		}
		pattern.Files[file.Name()] = string(data)
	}

	return nil
}

// initializeDefaultPatterns initializes the repository with default patterns
func (r *PatternRepository) initializeDefaultPatterns() error {
	// Create default patterns directory structure
	defaultPatterns := []*Pattern{
		{
			ID:          "aws-vpc-basic",
			Name:        "AWS VPC Basic",
			Description: "A basic AWS VPC with public and private subnets",
			Category:    CategoryNetworking,
			Provider:    ProviderAWS,
			Complexity:  ComplexityBasic,
			Tags:        []string{"vpc", "networking", "aws"},
			Files: map[string]string{
				"main.tf": defaultAWSVPCMainTF,
				"variables.tf": defaultAWSVPCVariablesTF,
				"outputs.tf": defaultAWSVPCOutputsTF,
				"README.md": defaultAWSVPCReadme,
			},
		},
		{
			ID:          "aws-ec2-web-server",
			Name:        "AWS EC2 Web Server",
			Description: "EC2 instance with web server configuration",
			Category:    CategoryCompute,
			Provider:    ProviderAWS,
			Complexity:  ComplexityIntermediate,
			Tags:        []string{"ec2", "web", "compute", "aws"},
			Files: map[string]string{
				"main.tf": defaultAWSEC2MainTF,
				"variables.tf": defaultAWSEC2VariablesTF,
				"outputs.tf": defaultAWSEC2OutputsTF,
				"README.md": defaultAWSEC2Readme,
			},
		},
		{
			ID:          "azure-vnet-basic",
			Name:        "Azure VNet Basic",
			Description: "A basic Azure Virtual Network with subnets",
			Category:    CategoryNetworking,
			Provider:    ProviderAzure,
			Complexity:  ComplexityBasic,
			Tags:        []string{"vnet", "networking", "azure"},
			Files: map[string]string{
				"main.tf": defaultAzureVNetMainTF,
				"variables.tf": defaultAzureVNetVariablesTF,
				"outputs.tf": defaultAzureVNetOutputsTF,
				"README.md": defaultAzureVNetReadme,
			},
		},
		{
			ID:          "gcp-vpc-basic",
			Name:        "GCP VPC Basic",
			Description: "A basic GCP VPC with subnets",
			Category:    CategoryNetworking,
			Provider:    ProviderGCP,
			Complexity:  ComplexityBasic,
			Tags:        []string{"vpc", "networking", "gcp"},
			Files: map[string]string{
				"main.tf": defaultGCPVPCMainTF,
				"variables.tf": defaultGCPVPCVariablesTF,
				"outputs.tf": defaultGCPVPCOutputsTF,
				"README.md": defaultGCPVPCReadme,
			},
		},
		{
			ID:          "terraform-module-structure",
			Name:        "Terraform Module Structure",
			Description: "Standard structure for a Terraform module",
			Category:    CategoryApplication,
			Provider:    ProviderGeneric,
			Complexity:  ComplexityBasic,
			Tags:        []string{"module", "structure", "template"},
			Files: map[string]string{
				"main.tf": defaultModuleMainTF,
				"variables.tf": defaultModuleVariablesTF,
				"outputs.tf": defaultModuleOutputsTF,
				"README.md": defaultModuleReadme,
				".gitignore": defaultGitignore,
			},
		},
	}

	// Create index file
	indexData, err := json.MarshalIndent(defaultPatterns, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default patterns: %w", err)
	}

	indexPath := filepath.Join(r.patternPath, "index.json")
	if err := ioutil.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("failed to write pattern index: %w", err)
	}

	// Create pattern directories and files
	for _, pattern := range defaultPatterns {
		patternDir := filepath.Join(r.patternPath, pattern.ID)
		if err := os.MkdirAll(patternDir, 0755); err != nil {
			return fmt.Errorf("failed to create pattern directory: %w", err)
		}

		for fileName, fileContent := range pattern.Files {
			filePath := filepath.Join(patternDir, fileName)
			if err := ioutil.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
				return fmt.Errorf("failed to write pattern file: %w", err)
			}
		}

		r.patterns[pattern.ID] = pattern
	}

	return nil
}

// GetPatternByID returns a pattern by ID
func (r *PatternRepository) GetPatternByID(id string) (*Pattern, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	pattern, ok := r.patterns[id]
	if !ok {
		return nil, fmt.Errorf("pattern not found: %s", id)
	}

	return pattern, nil
}

// FindPatterns returns patterns matching the filter criteria
func (r *PatternRepository) FindPatterns(filter PatternFilter) ([]*Pattern, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var results []*Pattern

	for _, pattern := range r.patterns {
		// Filter by category
		if filter.Category != nil && pattern.Category != *filter.Category {
			continue
		}

		// Filter by provider
		if filter.Provider != nil && pattern.Provider != *filter.Provider {
			continue
		}

		// Filter by complexity
		if filter.Complexity != nil && pattern.Complexity != *filter.Complexity {
			continue
		}

		// Filter by tags
		if len(filter.Tags) > 0 {
			matches := false
			for _, tag := range filter.Tags {
				for _, patternTag := range pattern.Tags {
					if strings.EqualFold(tag, patternTag) {
						matches = true
						break
					}
				}
				if matches {
					break
				}
			}
			if !matches {
				continue
			}
		}

		// Filter by query
		if filter.Query != "" {
			query := strings.ToLower(filter.Query)
			name := strings.ToLower(pattern.Name)
			desc := strings.ToLower(pattern.Description)
			id := strings.ToLower(pattern.ID)

			if !strings.Contains(name, query) && !strings.Contains(desc, query) && !strings.Contains(id, query) {
				continue
			}
		}

		results = append(results, pattern)
	}

	return results, nil
}

// Default pattern templates
const (
	defaultAWSVPCMainTF = `# main.tf
# AWS VPC Module - Main Configuration
# This module creates a VPC with public and private subnets across multiple AZs.

provider "aws" {
  region = var.region
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(
    var.tags,
    {
      Name = var.name
    }
  )
}

# Public subnets
resource "aws_subnet" "public" {
  count             = length(var.public_subnets)
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.public_subnets[count.index]
  availability_zone = var.availability_zones[count.index % length(var.availability_zones)]
  
  map_public_ip_on_launch = true

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-public-${count.index + 1}"
      Tier = "Public"
    }
  )
}

# Private subnets
resource "aws_subnet" "private" {
  count             = length(var.private_subnets)
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnets[count.index]
  availability_zone = var.availability_zones[count.index % length(var.availability_zones)]

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-private-${count.index + 1}"
      Tier = "Private"
    }
  )
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-igw"
    }
  )
}

# Elastic IP for NAT Gateway
resource "aws_eip" "nat" {
  count = length(var.public_subnets) > 0 ? 1 : 0
  
  domain = "vpc"

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-nat-eip"
    }
  )
}

# NAT Gateway
resource "aws_nat_gateway" "main" {
  count = length(var.public_subnets) > 0 && length(var.private_subnets) > 0 ? 1 : 0
  
  allocation_id = aws_eip.nat[0].id
  subnet_id     = aws_subnet.public[0].id

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-nat"
    }
  )

  depends_on = [aws_internet_gateway.main]
}

# Route Tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-public-rt"
    }
  )
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_route_table" "private" {
  count  = length(var.private_subnets) > 0 ? 1 : 0
  vpc_id = aws_vpc.main.id

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-private-rt"
    }
  )
}

resource "aws_route" "private_nat_gateway" {
  count                  = length(var.private_subnets) > 0 ? 1 : 0
  route_table_id         = aws_route_table.private[0].id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.main[0].id
}

# Route Table Associations
resource "aws_route_table_association" "public" {
  count          = length(var.public_subnets)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count          = length(var.private_subnets)
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[0].id
}
`

	defaultAWSVPCVariablesTF = `# variables.tf
# AWS VPC Module - Variables

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "name" {
  description = "Name to be used on all the resources as identifier"
  type        = string
}

variable "vpc_cidr" {
  description = "The CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "A list of availability zones in the region"
  type        = list(string)
  default     = []
}

variable "public_subnets" {
  description = "A list of public subnets CIDR blocks inside the VPC"
  type        = list(string)
  default     = []
}

variable "private_subnets" {
  description = "A list of private subnets CIDR blocks inside the VPC"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
`

	defaultAWSVPCOutputsTF = `# outputs.tf
# AWS VPC Module - Outputs

output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.main.id
}

output "vpc_cidr_block" {
  description = "The CIDR block of the VPC"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "List of IDs of private subnets"
  value       = aws_subnet.private[*].id
}

output "public_route_table_id" {
  description = "ID of the public route table"
  value       = aws_route_table.public.id
}

output "private_route_table_ids" {
  description = "List of IDs of private route tables"
  value       = aws_route_table.private[*].id
}

output "nat_gateway_ids" {
  description = "List of NAT Gateway IDs"
  value       = aws_nat_gateway.main[*].id
}

output "internet_gateway_id" {
  description = "ID of the Internet Gateway"
  value       = aws_internet_gateway.main.id
}
`

	defaultAWSVPCReadme = `# AWS VPC Terraform Module

This module creates a VPC with public and private subnets across multiple Availability Zones.

## Usage

```hcl
module "vpc" {
  source = "./path/to/module"

  name = "my-vpc"
  vpc_cidr = "10.0.0.0/16"

  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  private_subnets    = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  tags = {
    Environment = "production"
    Project     = "networking"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | >= 4.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| name | Name to be used on all the resources as identifier | `string` | n/a | yes |
| vpc_cidr | The CIDR block for the VPC | `string` | `"10.0.0.0/16"` | no |
| availability_zones | A list of availability zones in the region | `list(string)` | `[]` | no |
| public_subnets | A list of public subnets CIDR blocks inside the VPC | `list(string)` | `[]` | no |
| private_subnets | A list of private subnets CIDR blocks inside the VPC | `list(string)` | `[]` | no |
| tags | A map of tags to add to all resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| vpc_id | The ID of the VPC |
| vpc_cidr_block | The CIDR block of the VPC |
| public_subnet_ids | List of IDs of public subnets |
| private_subnet_ids | List of IDs of private subnets |
| public_route_table_id | ID of the public route table |
| private_route_table_ids | List of IDs of private route tables |
| nat_gateway_ids | List of NAT Gateway IDs |
| internet_gateway_id | ID of the Internet Gateway |

## Best Practices Followed

1. **Resource Organization**: Logically organized resources by type and function
2. **Naming Convention**: Consistent naming with prefixes for all resources
3. **Tagging Strategy**: Comprehensive tagging for resource management
4. **Modular Design**: Components can be enabled/disabled based on inputs
5. **Variable Validation**: Clear variable descriptions and types
6. **Output Documentation**: Comprehensive outputs with descriptions
7. **Security Considerations**: Public and private subnet separation
`

	defaultAWSEC2MainTF = `# EC2 Instance with Web Server
# This module creates an EC2 instance with web server configuration

provider "aws" {
  region = var.region
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_security_group" "web" {
  name        = "${var.name}-web-sg"
  description = "Security group for web server"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow HTTP traffic"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow HTTPS traffic"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.ssh_allowed_cidr_blocks
    description = "Allow SSH access from specified CIDRs"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-web-sg"
    }
  )
}

resource "aws_instance" "web" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [aws_security_group.web.id]
  iam_instance_profile   = var.iam_instance_profile
  key_name               = var.key_name

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size
    delete_on_termination = true
    encrypted             = true
  }

  user_data = <<-EOF
    #!/bin/bash
    yum update -y
    yum install -y httpd
    systemctl start httpd
    systemctl enable httpd
    echo "<h1>Web Server deployed with Terraform</h1>" > /var/www/html/index.html
    echo "<p>Server ID: $(hostname)</p>" >> /var/www/html/index.html
  EOF

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-web-server"
    }
  )

  volume_tags = merge(
    var.tags,
    {
      Name = "${var.name}-web-server-ebs"
    }
  )
}

resource "aws_eip" "web" {
  count  = var.assign_eip ? 1 : 0
  domain = "vpc"
  instance = aws_instance.web.id

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-web-eip"
    }
  )
}
`

	defaultAWSEC2VariablesTF = `# EC2 Instance with Web Server - Variables

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "name" {
  description = "Name prefix for resources"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC where resources will be created"
  type        = string
}

variable "subnet_id" {
  description = "ID of the subnet where EC2 instance will be created"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "root_volume_size" {
  description = "Size of the root EBS volume in GB"
  type        = number
  default     = 20
}

variable "assign_eip" {
  description = "Whether to assign an Elastic IP to the instance"
  type        = bool
  default     = true
}

variable "key_name" {
  description = "Name of the SSH key pair to use"
  type        = string
  default     = null
}

variable "iam_instance_profile" {
  description = "IAM instance profile name for the EC2 instance"
  type        = string
  default     = null
}

variable "ssh_allowed_cidr_blocks" {
  description = "List of CIDR blocks allowed to SSH to the instance"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
`

	defaultAWSEC2OutputsTF = `# EC2 Instance with Web Server - Outputs

output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.web.id
}

output "instance_arn" {
  description = "ARN of the EC2 instance"
  value       = aws_instance.web.arn
}

output "instance_public_ip" {
  description = "Public IP of the EC2 instance (if EIP is enabled)"
  value       = var.assign_eip ? aws_eip.web[0].public_ip : aws_instance.web.public_ip
}

output "instance_private_ip" {
  description = "Private IP of the EC2 instance"
  value       = aws_instance.web.private_ip
}

output "security_group_id" {
  description = "ID of the security group attached to the instance"
  value       = aws_security_group.web.id
}

output "eip_id" {
  description = "ID of the Elastic IP (if enabled)"
  value       = var.assign_eip ? aws_eip.web[0].id : null
}
`

	defaultAWSEC2Readme = `# AWS EC2 Web Server Module

This module creates an EC2 instance configured as a web server with appropriate security group rules.

## Usage

```hcl
module "web_server" {
  source = "./path/to/module"

  name      = "app"
  vpc_id    = "vpc-12345678"
  subnet_id = "subnet-12345678"
  
  instance_type = "t3.small"
  key_name      = "my-key-pair"
  
  ssh_allowed_cidr_blocks = ["10.0.0.0/8"]
  
  tags = {
    Environment = "production"
    Application = "web"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | >= 4.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| name | Name prefix for resources | `string` | n/a | yes |
| vpc_id | ID of the VPC where resources will be created | `string` | n/a | yes |
| subnet_id | ID of the subnet where EC2 instance will be created | `string` | n/a | yes |
| instance_type | EC2 instance type | `string` | `"t3.micro"` | no |
| root_volume_size | Size of the root EBS volume in GB | `number` | `20` | no |
| assign_eip | Whether to assign an Elastic IP to the instance | `bool` | `true` | no |
| key_name | Name of the SSH key pair to use | `string` | `null` | no |
| iam_instance_profile | IAM instance profile name for the EC2 instance | `string` | `null` | no |
| ssh_allowed_cidr_blocks | List of CIDR blocks allowed to SSH to the instance | `list(string)` | `["0.0.0.0/0"]` | no |
| tags | A map of tags to add to all resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| instance_id | ID of the EC2 instance |
| instance_arn | ARN of the EC2 instance |
| instance_public_ip | Public IP of the EC2 instance (if EIP is enabled) |
| instance_private_ip | Private IP of the EC2 instance |
| security_group_id | ID of the security group attached to the instance |
| eip_id | ID of the Elastic IP (if enabled) |

## Best Practices Followed

1. **Security Groups**: Least privilege access with specific ingress/egress rules
2. **Root EBS Volumes**: Encrypted by default with suitable volume type (gp3)
3. **AMI Selection**: Using official Amazon Linux 2 AMI with dynamic selection
4. **Resource Naming**: Consistent naming with prefixes for all resources
5. **Tagging Strategy**: Comprehensive tagging for resource management
6. **Network Isolation**: Configurable subnet placement for proper network isolation
7. **SSH Access Control**: Limited SSH access with configurable CIDR blocks
`

	defaultAzureVNetMainTF = `# Azure Virtual Network Module - Main Configuration
# This module creates a Virtual Network with subnets in Azure

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "this" {
  count    = var.create_resource_group ? 1 : 0
  name     = var.resource_group_name
  location = var.location

  tags = var.tags
}

locals {
  resource_group_name = var.create_resource_group ? azurerm_resource_group.this[0].name : var.resource_group_name
}

resource "azurerm_virtual_network" "this" {
  name                = var.vnet_name
  resource_group_name = local.resource_group_name
  location            = var.location
  address_space       = var.address_space
  dns_servers         = var.dns_servers
  tags                = var.tags
}

resource "azurerm_subnet" "this" {
  for_each = var.subnets

  name                 = each.key
  resource_group_name  = local.resource_group_name
  virtual_network_name = azurerm_virtual_network.this.name
  address_prefixes     = [each.value.address_prefix]
  
  service_endpoints    = lookup(each.value, "service_endpoints", null)
  
  dynamic "delegation" {
    for_each = lookup(each.value, "delegation", {}) != {} ? [1] : []
    
    content {
      name = lookup(each.value.delegation, "name", null)
      
      service_delegation {
        name    = lookup(each.value.delegation.service_delegation, "name", null)
        actions = lookup(each.value.delegation.service_delegation, "actions", null)
      }
    }
  }
}

resource "azurerm_network_security_group" "this" {
  for_each = var.network_security_groups

  name                = each.key
  resource_group_name = local.resource_group_name
  location            = var.location
  tags                = var.tags
}

resource "azurerm_subnet_network_security_group_association" "this" {
  for_each = {
    for k, v in var.network_security_groups : k => v
    if contains(keys(var.subnets), v.subnet_name)
  }

  subnet_id                 = azurerm_subnet.this[each.value.subnet_name].id
  network_security_group_id = azurerm_network_security_group.this[each.key].id
}

resource "azurerm_network_security_rule" "this" {
  for_each = var.network_security_rules

  name                         = each.key
  resource_group_name          = local.resource_group_name
  network_security_group_name  = each.value.network_security_group_name
  priority                     = each.value.priority
  direction                    = each.value.direction
  access                       = each.value.access
  protocol                     = each.value.protocol
  source_port_range            = lookup(each.value, "source_port_range", "*")
  destination_port_range       = lookup(each.value, "destination_port_range", "*")
  source_address_prefix        = lookup(each.value, "source_address_prefix", "*")
  destination_address_prefix   = lookup(each.value, "destination_address_prefix", "*")
  source_port_ranges           = lookup(each.value, "source_port_ranges", null)
  destination_port_ranges      = lookup(each.value, "destination_port_ranges", null)
  source_address_prefixes      = lookup(each.value, "source_address_prefixes", null)
  destination_address_prefixes = lookup(each.value, "destination_address_prefixes", null)
}

resource "azurerm_route_table" "this" {
  for_each = var.route_tables

  name                = each.key
  resource_group_name = local.resource_group_name
  location            = var.location
  tags                = var.tags
}

resource "azurerm_route" "this" {
  for_each = var.routes

  name                   = each.key
  resource_group_name    = local.resource_group_name
  route_table_name       = each.value.route_table_name
  address_prefix         = each.value.address_prefix
  next_hop_type          = each.value.next_hop_type
  next_hop_in_ip_address = lookup(each.value, "next_hop_in_ip_address", null)
}

resource "azurerm_subnet_route_table_association" "this" {
  for_each = {
    for k, v in var.route_tables : k => v
    if contains(keys(var.subnets), v.subnet_name)
  }

  subnet_id      = azurerm_subnet.this[each.value.subnet_name].id
  route_table_id = azurerm_route_table.this[each.key].id
}
`

	defaultAzureVNetVariablesTF = `# Azure Virtual Network Module - Variables

variable "create_resource_group" {
  description = "Controls if the resource group should be created"
  type        = bool
  default     = false
}

variable "resource_group_name" {
  description = "The name of the resource group to use"
  type        = string
}

variable "location" {
  description = "The Azure region where resources will be created"
  type        = string
}

variable "vnet_name" {
  description = "The name of the virtual network"
  type        = string
}

variable "address_space" {
  description = "The address space for the virtual network"
  type        = list(string)
}

variable "dns_servers" {
  description = "List of DNS servers to use for the VNet"
  type        = list(string)
  default     = []
}

variable "subnets" {
  description = "Map of subnet objects. Key is subnet name, value is subnet configuration."
  type        = map(any)
  default     = {}
}

variable "network_security_groups" {
  description = "Map of network security groups to create. Key is NSG name, value is NSG configuration."
  type        = map(any)
  default     = {}
}

variable "network_security_rules" {
  description = "Map of network security rules to create. Key is rule name, value is rule configuration."
  type        = map(any)
  default     = {}
}

variable "route_tables" {
  description = "Map of route tables to create. Key is route table name, value is route table configuration."
  type        = map(any)
  default     = {}
}

variable "routes" {
  description = "Map of routes to create. Key is route name, value is route configuration."
  type        = map(any)
  default     = {}
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
`

	defaultAzureVNetOutputsTF = `# Azure Virtual Network Module - Outputs

output "vnet_id" {
  description = "The ID of the virtual network"
  value       = azurerm_virtual_network.this.id
}

output "vnet_name" {
  description = "The name of the virtual network"
  value       = azurerm_virtual_network.this.name
}

output "vnet_address_space" {
  description = "The address space of the virtual network"
  value       = azurerm_virtual_network.this.address_space
}

output "subnet_ids" {
  description = "Map of subnet names to subnet IDs"
  value       = { for k, v in azurerm_subnet.this : k => v.id }
}

output "subnet_address_prefixes" {
  description = "Map of subnet names to subnet address prefixes"
  value       = { for k, v in azurerm_subnet.this : k => v.address_prefixes }
}

output "network_security_group_ids" {
  description = "Map of network security group names to NSG IDs"
  value       = { for k, v in azurerm_network_security_group.this : k => v.id }
}

output "route_table_ids" {
  description = "Map of route table names to route table IDs"
  value       = { for k, v in azurerm_route_table.this : k => v.id }
}

output "resource_group_name" {
  description = "The name of the resource group"
  value       = local.resource_group_name
}
`

	defaultAzureVNetReadme = `# Azure Virtual Network Terraform Module

This module creates a Virtual Network with subnets in Azure, along with associated network security groups and route tables.

## Usage

```hcl
module "vnet" {
  source = "./path/to/module"

  vnet_name           = "example-vnet"
  resource_group_name = "example-rg"
  location            = "eastus"
  address_space       = ["10.0.0.0/16"]

  subnets = {
    web = {
      address_prefix    = "10.0.1.0/24"
      service_endpoints = ["Microsoft.Storage", "Microsoft.Sql"]
    }
    app = {
      address_prefix = "10.0.2.0/24"
    }
    db = {
      address_prefix = "10.0.3.0/24"
      delegation = {
        name = "delegation"
        service_delegation = {
          name    = "Microsoft.Sql/servers"
          actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
        }
      }
    }
  }

  network_security_groups = {
    web-nsg = {
      subnet_name = "web"
    }
    app-nsg = {
      subnet_name = "app"
    }
  }

  network_security_rules = {
    web-allow-http = {
      network_security_group_name = "web-nsg"
      priority                     = 100
      direction                    = "Inbound"
      access                       = "Allow"
      protocol                     = "Tcp"
      destination_port_range       = "80"
    }
    web-allow-https = {
      network_security_group_name = "web-nsg"
      priority                     = 110
      direction                    = "Inbound"
      access                       = "Allow"
      protocol                     = "Tcp"
      destination_port_range       = "443"
    }
  }

  tags = {
    Environment = "Production"
    Project     = "Example"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| azurerm | >= 3.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| create_resource_group | Controls if the resource group should be created | `bool` | `false` | no |
| resource_group_name | The name of the resource group to use | `string` | n/a | yes |
| location | The Azure region where resources will be created | `string` | n/a | yes |
| vnet_name | The name of the virtual network | `string` | n/a | yes |
| address_space | The address space for the virtual network | `list(string)` | n/a | yes |
| dns_servers | List of DNS servers to use for the VNet | `list(string)` | `[]` | no |
| subnets | Map of subnet objects | `map(any)` | `{}` | no |
| network_security_groups | Map of network security groups to create | `map(any)` | `{}` | no |
| network_security_rules | Map of network security rules to create | `map(any)` | `{}` | no |
| route_tables | Map of route tables to create | `map(any)` | `{}` | no |
| routes | Map of routes to create | `map(any)` | `{}` | no |
| tags | A map of tags to add to all resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| vnet_id | The ID of the virtual network |
| vnet_name | The name of the virtual network |
| vnet_address_space | The address space of the virtual network |
| subnet_ids | Map of subnet names to subnet IDs |
| subnet_address_prefixes | Map of subnet names to subnet address prefixes |
| network_security_group_ids | Map of network security group names to NSG IDs |
| route_table_ids | Map of route table names to route table IDs |
| resource_group_name | The name of the resource group |

## Best Practices Followed

1. **Modularity**: Network components organized in a modular way for reuse
2. **Resource Naming**: Consistent naming with clear identifiers
3. **Security**: Network security groups integrated with subnets
4. **Delegation Support**: Subnet delegation for PaaS services
5. **Tagging Strategy**: Comprehensive tagging for all resources
6. **Local Variables**: Using locals to simplify resource referencing
7. **Flexible Configuration**: Map-based inputs for easy customization
8. **Output Documentation**: Detailed output descriptions for resource references
`

	defaultGCPVPCMainTF = `# GCP VPC Module - Main Configuration
# This module creates a VPC with subnets in Google Cloud Platform

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_compute_network" "vpc" {
  name                    = var.name
  auto_create_subnetworks = var.auto_create_subnetworks
  routing_mode            = var.routing_mode
  description             = var.description
  delete_default_routes_on_create = var.delete_default_routes_on_create
}

resource "google_compute_subnetwork" "subnets" {
  for_each = { for subnet in var.subnets : subnet.name => subnet }

  name                     = each.value.name
  network                  = google_compute_network.vpc.id
  ip_cidr_range            = each.value.ip_cidr_range
  region                   = lookup(each.value, "region", var.region)
  private_ip_google_access = lookup(each.value, "private_ip_google_access", true)
  description              = lookup(each.value, "description", null)

  dynamic "secondary_ip_range" {
    for_each = lookup(each.value, "secondary_ip_ranges", [])
    content {
      range_name    = secondary_ip_range.value.range_name
      ip_cidr_range = secondary_ip_range.value.ip_cidr_range
    }
  }

  log_config {
    aggregation_interval = lookup(each.value, "log_config_aggregation_interval", "INTERVAL_5_SEC")
    flow_sampling        = lookup(each.value, "log_config_flow_sampling", 0.5)
    metadata             = lookup(each.value, "log_config_metadata", "INCLUDE_ALL_METADATA")
  }
}

resource "google_compute_firewall" "rules" {
  for_each = { for rule in var.firewall_rules : rule.name => rule }

  name        = each.value.name
  network     = google_compute_network.vpc.id
  description = lookup(each.value, "description", null)
  direction   = lookup(each.value, "direction", "INGRESS")
  priority    = lookup(each.value, "priority", 1000)

  dynamic "allow" {
    for_each = lookup(each.value, "allow", [])
    content {
      protocol = allow.value.protocol
      ports    = lookup(allow.value, "ports", null)
    }
  }

  dynamic "deny" {
    for_each = lookup(each.value, "deny", [])
    content {
      protocol = deny.value.protocol
      ports    = lookup(deny.value, "ports", null)
    }
  }

  source_tags             = lookup(each.value, "source_tags", null)
  source_ranges           = lookup(each.value, "source_ranges", null)
  source_service_accounts = lookup(each.value, "source_service_accounts", null)
  target_tags             = lookup(each.value, "target_tags", null)
  target_service_accounts = lookup(each.value, "target_service_accounts", null)
}

resource "google_compute_router" "router" {
  count   = var.create_router ? 1 : 0
  name    = "${var.name}-router"
  network = google_compute_network.vpc.id
  region  = var.region

  bgp {
    asn = var.router_asn
  }
}

resource "google_compute_router_nat" "nat" {
  count                              = var.create_router && var.create_nat ? 1 : 0
  name                               = "${var.name}-nat"
  router                             = google_compute_router.router[0].name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

resource "google_compute_route" "routes" {
  for_each = { for route in var.routes : route.name => route }

  name             = each.value.name
  network          = google_compute_network.vpc.id
  dest_range       = each.value.dest_range
  priority         = lookup(each.value, "priority", 1000)
  description      = lookup(each.value, "description", null)
  
  next_hop_gateway           = lookup(each.value, "next_hop_gateway", null)
  next_hop_instance          = lookup(each.value, "next_hop_instance", null)
  next_hop_instance_zone     = lookup(each.value, "next_hop_instance_zone", null)
  next_hop_ip                = lookup(each.value, "next_hop_ip", null)
  next_hop_vpn_tunnel        = lookup(each.value, "next_hop_vpn_tunnel", null)
  next_hop_ilb               = lookup(each.value, "next_hop_ilb", null)
}
`

	defaultGCPVPCVariablesTF = `# GCP VPC Module - Variables

variable "project_id" {
  description = "The ID of the project where resources will be created"
  type        = string
}

variable "region" {
  description = "The region where resources will be created"
  type        = string
}

variable "name" {
  description = "The name of the VPC network"
  type        = string
}

variable "auto_create_subnetworks" {
  description = "When set to true, the network is created in auto subnet mode"
  type        = bool
  default     = false
}

variable "routing_mode" {
  description = "The network routing mode (REGIONAL or GLOBAL)"
  type        = string
  default     = "GLOBAL"
}

variable "description" {
  description = "Description of the VPC network"
  type        = string
  default     = "Managed by Terraform"
}

variable "delete_default_routes_on_create" {
  description = "If set to true, default routes (0.0.0.0/0) will be deleted immediately after network creation"
  type        = bool
  default     = false
}

variable "subnets" {
  description = "The list of subnets to create within the VPC"
  type        = list(object({
    name                     = string
    ip_cidr_range            = string
    region                   = optional(string)
    private_ip_google_access = optional(bool, true)
    description              = optional(string)
    secondary_ip_ranges      = optional(list(object({
      range_name    = string
      ip_cidr_range = string
    })), [])
    log_config_aggregation_interval = optional(string, "INTERVAL_5_SEC")
    log_config_flow_sampling        = optional(number, 0.5)
    log_config_metadata             = optional(string, "INCLUDE_ALL_METADATA")
  }))
  default     = []
}

variable "firewall_rules" {
  description = "List of firewall rules to create"
  type        = list(any)
  default     = []
}

variable "routes" {
  description = "List of routes to create"
  type        = list(any)
  default     = []
}

variable "create_router" {
  description = "Whether to create a Cloud Router"
  type        = bool
  default     = false
}

variable "create_nat" {
  description = "Whether to create a Cloud NAT gateway"
  type        = bool
  default     = false
}

variable "router_asn" {
  description = "ASN for the Cloud Router"
  type        = number
  default     = 64514
}
`

	defaultGCPVPCOutputsTF = `# GCP VPC Module - Outputs

output "vpc_id" {
  description = "The ID of the VPC"
  value       = google_compute_network.vpc.id
}

output "vpc_name" {
  description = "The name of the VPC"
  value       = google_compute_network.vpc.name
}

output "vpc_self_link" {
  description = "The URI of the VPC"
  value       = google_compute_network.vpc.self_link
}

output "subnet_ids" {
  description = "Map of subnet names to subnet IDs"
  value       = { for name, subnet in google_compute_subnetwork.subnets : name => subnet.id }
}

output "subnet_self_links" {
  description = "Map of subnet names to subnet self links"
  value       = { for name, subnet in google_compute_subnetwork.subnets : name => subnet.self_link }
}

output "subnet_ip_cidr_ranges" {
  description = "Map of subnet names to primary IP CIDR ranges"
  value       = { for name, subnet in google_compute_subnetwork.subnets : name => subnet.ip_cidr_range }
}

output "subnet_secondary_ranges" {
  description = "Map of subnet names to a list of secondary IP range names and ranges"
  value       = { for name, subnet in google_compute_subnetwork.subnets : name => subnet.secondary_ip_range }
}

output "router_id" {
  description = "The ID of the Cloud Router (if created)"
  value       = var.create_router ? google_compute_router.router[0].id : null
}

output "router_self_link" {
  description = "The URI of the Cloud Router (if created)"
  value       = var.create_router ? google_compute_router.router[0].self_link : null
}

output "nat_id" {
  description = "The ID of the Cloud NAT (if created)"
  value       = var.create_router && var.create_nat ? google_compute_router_nat.nat[0].id : null
}

output "nat_name" {
  description = "The name of the Cloud NAT (if created)"
  value       = var.create_router && var.create_nat ? google_compute_router_nat.nat[0].name : null
}
`

	defaultGCPVPCReadme = `# GCP VPC Terraform Module

This module creates a VPC network in Google Cloud Platform with subnets, firewall rules, Cloud Router, and Cloud NAT.

## Usage

```hcl
module "vpc" {
  source = "./path/to/module"

  project_id = "my-project"
  region     = "us-central1"
  name       = "my-vpc"

  subnets = [
    {
      name          = "subnet-01"
      ip_cidr_range = "10.10.10.0/24"
      region        = "us-central1"
      secondary_ip_ranges = [
        {
          range_name    = "pods"
          ip_cidr_range = "10.20.0.0/22"
        },
        {
          range_name    = "services"
          ip_cidr_range = "10.30.0.0/24"
        }
      ]
    },
    {
      name          = "subnet-02"
      ip_cidr_range = "10.10.20.0/24"
      region        = "us-east1"
    }
  ]

  firewall_rules = [
    {
      name        = "allow-internal"
      description = "Allow internal traffic"
      direction   = "INGRESS"
      priority    = 1000
      source_ranges = ["10.10.10.0/24", "10.10.20.0/24"]
      allow = [
        {
          protocol = "tcp"
        },
        {
          protocol = "udp"
        },
        {
          protocol = "icmp"
        }
      ]
    },
    {
      name        = "allow-ssh"
      description = "Allow SSH from specific IPs"
      direction   = "INGRESS"
      priority    = 1000
      source_ranges = ["35.235.240.0/20"]
      allow = [
        {
          protocol = "tcp"
          ports    = ["22"]
        }
      ]
    }
  ]

  create_router = true
  create_nat    = true
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| google | >= 4.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| project_id | The ID of the project where resources will be created | `string` | n/a | yes |
| region | The region where resources will be created | `string` | n/a | yes |
| name | The name of the VPC network | `string` | n/a | yes |
| auto_create_subnetworks | When set to true, the network is created in auto subnet mode | `bool` | `false` | no |
| routing_mode | The network routing mode (REGIONAL or GLOBAL) | `string` | `"GLOBAL"` | no |
| description | Description of the VPC network | `string` | `"Managed by Terraform"` | no |
| delete_default_routes_on_create | If set to true, default routes (0.0.0.0/0) will be deleted immediately after network creation | `bool` | `false` | no |
| subnets | The list of subnets to create within the VPC | `list(object)` | `[]` | no |
| firewall_rules | List of firewall rules to create | `list(any)` | `[]` | no |
| routes | List of routes to create | `list(any)` | `[]` | no |
| create_router | Whether to create a Cloud Router | `bool` | `false` | no |
| create_nat | Whether to create a Cloud NAT gateway | `bool` | `false` | no |
| router_asn | ASN for the Cloud Router | `number` | `64514` | no |

## Outputs

| Name | Description |
|------|-------------|
| vpc_id | The ID of the VPC |
| vpc_name | The name of the VPC |
| vpc_self_link | The URI of the VPC |
| subnet_ids | Map of subnet names to subnet IDs |
| subnet_self_links | Map of subnet names to subnet self links |
| subnet_ip_cidr_ranges | Map of subnet names to primary IP CIDR ranges |
| subnet_secondary_ranges | Map of subnet names to a list of secondary IP range names and ranges |
| router_id | The ID of the Cloud Router (if created) |
| router_self_link | The URI of the Cloud Router (if created) |
| nat_id | The ID of the Cloud NAT (if created) |
| nat_name | The name of the Cloud NAT (if created) |

## Best Practices Followed

1. **Modular Design**: Components organized for independent management
2. **Network Segmentation**: Proper subnet organization and CIDR planning
3. **Security Controls**: Firewall rules with least privilege access
4. **Private Google Access**: Enabled by default for secure API access
5. **Flow Logs**: Enabled by default for network monitoring and debugging
6. **Secondary IP Ranges**: Support for GKE clusters and other services
7. **Cloud NAT**: Optional egress-only internet access for private instances
8. **Flexible Configuration**: Input variables with reasonable defaults
`

	defaultModuleMainTF = `# Terraform Module - Main Configuration
# This is the main configuration file for a Terraform module.
# It should contain the core resource definitions and logic.

# Main resources and logic go here

# Example resource block:
# resource "example_resource" "main" {
#   name = var.name
#   # other attributes
# }

# For child modules, use:
# module "example" {
#   source = "./modules/example"
#   # input variables
# }

# Data lookup example:
# data "example_data" "lookup" {
#   name = var.lookup_name
# }

# Local variables:
locals {
  # Define local variables to simplify expressions or for reuse
  # common_tags = merge(var.tags, {
  #   Module = "example"
  # })
}
`

	defaultModuleVariablesTF = `# Terraform Module - Variables
# This file contains variable definitions for the module.
# Each variable should include a description and type.
# Default values should be provided where appropriate.

# Required variables (no default value)
variable "name" {
  description = "Name to be used for resources created by this module"
  type        = string
}

# Optional variables (with default values)
variable "enabled" {
  description = "Whether resources in this module should be created"
  type        = bool
  default     = true
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

# Complex type example
variable "config" {
  description = "Configuration options for the module"
  type = object({
    option_a = string
    option_b = number
    option_c = bool
    nested = object({
      sub_option = string
    })
  })
  default = {
    option_a = "default"
    option_b = 123
    option_c = true
    nested = {
      sub_option = "default"
    }
  }
}

# List/set example
variable "allowed_ips" {
  description = "List of allowed IP addresses"
  type        = list(string)
  default     = []
}

# Variable with validation
variable "environment" {
  description = "Environment where resources will be deployed"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}
`

	defaultModuleOutputsTF = `# Terraform Module - Outputs
# This file contains output definitions for the module.
# Each output should include a description.

output "id" {
  description = "The ID of the main resource created by the module"
  value       = var.enabled ? resource.example.main.id : null
}

output "arn" {
  description = "The ARN of the main resource created by the module"
  value       = var.enabled ? resource.example.main.arn : null
}

output "name" {
  description = "The name of the main resource created by the module"
  value       = var.name
}

# Output example with complex value
output "config_summary" {
  description = "Summary of the configuration used"
  value = {
    name    = var.name
    enabled = var.enabled
    options = var.config
  }
}

# Sensitive output example
output "sensitive_value" {
  description = "Sensitive value that should not be displayed in the UI"
  value       = "sensitive-data-here"
  sensitive   = true
}
`

	defaultModuleReadme = `# Terraform Module Template

This is a Terraform module template that follows best practices for module development.

## Features

- Follows HashiCorp's [Terraform Module Structure](https://developer.hashicorp.com/terraform/language/modules/develop/structure) guidelines
- Includes comprehensive variable validation
- Provides detailed documentation
- Implements consistent naming conventions
- Uses proper variable and output descriptions

## Usage

```hcl
module "example" {
  source = "./path/to/module"

  name = "example-resource"
  
  # Optional configuration
  enabled = true
  tags = {
    Environment = "production"
    Project     = "example"
  }
  
  config = {
    option_a = "custom-value"
    option_b = 456
    option_c = false
    nested = {
      sub_option = "custom-sub-value"
    }
  }
  
  allowed_ips = ["10.0.0.0/8", "192.168.1.0/24"]
  environment = "prod"
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0.0 |
| provider1 | >= 3.0.0 |
| provider2 | >= 2.0.0 |

## Providers

| Name | Version |
|------|---------|
| provider1 | >= 3.0.0 |
| provider2 | >= 2.0.0 |

## Resources

| Name | Type |
|------|------|
| [resource.example](https://registry.terraform.io/providers/example/latest/docs/resources/example) | resource |
| [data.example](https://registry.terraform.io/providers/example/latest/docs/data-sources/example) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| name | Name to be used for resources created by this module | `string` | n/a | yes |
| enabled | Whether resources in this module should be created | `bool` | `true` | no |
| tags | A map of tags to add to all resources | `map(string)` | `{}` | no |
| config | Configuration options for the module | `object({...})` | `{...}` | no |
| allowed_ips | List of allowed IP addresses | `list(string)` | `[]` | no |
| environment | Environment where resources will be deployed | `string` | `"dev"` | no |

## Outputs

| Name | Description |
|------|-------------|
| id | The ID of the main resource created by the module |
| arn | The ARN of the main resource created by the module |
| name | The name of the main resource created by the module |
| config_summary | Summary of the configuration used |
| sensitive_value | Sensitive value that should not be displayed in the UI |

## Best Practices Followed

1. **Consistent File Structure**: Following the standard Terraform module structure with main.tf, variables.tf, outputs.tf, and README.md
2. **Comprehensive Documentation**: Detailed README with examples, requirements, inputs, and outputs
3. **Input Variable Validation**: Using validation blocks to ensure inputs meet requirements
4. **Sensible Defaults**: Providing reasonable default values for optional variables
5. **Resource Tagging**: Supporting consistent resource tagging
6. **Conditional Creation**: Supporting enabling/disabling resource creation
7. **Output Documentation**: Detailed descriptions for all outputs
8. **Complex Type Support**: Using Terraform's complex types for structured inputs
9. **Consistent Naming**: Following naming conventions for all resources
10. **Sensitive Output Handling**: Marking sensitive outputs appropriately
`

	defaultGitignore = `# Local .terraform directories
**/.terraform/*

# .tfstate files
*.tfstate
*.tfstate.*

# Crash log files
crash.log
crash.*.log

# Exclude all .tfvars files, which are likely to contain sensitive data
*.tfvars
*.tfvars.json

# Ignore override files as they are usually used for local dev
override.tf
override.tf.json
*_override.tf
*_override.tf.json

# Ignore CLI configuration files
.terraformrc
terraform.rc

# Ignore plan output files
tfplan
*.tfplan

# Ignore .terraform.lock.hcl
.terraform.lock.hcl

# Ignore generated documentation
.terraform-docs.yml
docs/

# Ignore IDE and editor files
.idea/
.vscode/
*.swp
*.swo
*~

# OS specific files
.DS_Store
Thumbs.db
`
)

// DefaultAuthoritySources defines the default authority sources for Terraform best practices
var DefaultAuthoritySources = []string{
	"https://developer.hashicorp.com/terraform/language/modules/develop",
	"https://developer.hashicorp.com/terraform/language/style",
	"https://developer.hashicorp.com/validated-designs/terraform-operating-guides-adoption/terraform-workflows",
	"https://developer.hashicorp.com/terraform/tutorials/pro-cert/pro-review",
}
</content>
</invoke>