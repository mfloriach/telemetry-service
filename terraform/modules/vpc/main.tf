# VPC module: VPC, public + private subnets across 3 AZs, internet gateway, NAT gateways.

resource "aws_vpc" "this" {
  cidr_block           = var.cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.name_prefix}-vpc"
    Environment = var.name_prefix
  }
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = {
    Name        = "${var.name_prefix}-igw"
    Environment = var.name_prefix
  }
}

# Public subnets
resource "aws_subnet" "public" {
  count                   = length(var.availability_zones)
  vpc_id                  = aws_vpc.this.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.name_prefix}-public-${count.index + 1}"
    Environment = var.name_prefix
    Tier        = "public"
  }
}

# Private subnets
resource "aws_subnet" "private" {
  count                   = length(var.availability_zones)
  vpc_id                  = aws_vpc.this.id
  cidr_block              = var.private_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = false

  tags = {
    Name        = "${var.name_prefix}-private-${count.index + 1}"
    Environment = var.name_prefix
    Tier        = "private"
  }
}

# Elastic IPs for NAT gateways
resource "aws_eip" "nat" {
  count = var.enable_nat_gateway ? length(var.availability_zones) : 0
  vpc   = true

  tags = {
    Name        = "${var.name_prefix}-nat-eip-${count.index + 1}"
    Environment = var.name_prefix
  }
}

# NAT gateways (one per AZ for HA)
resource "aws_nat_gateway" "this" {
  count         = var.enable_nat_gateway ? length(var.availability_zones) : 0
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = {
    Name        = "${var.name_prefix}-nat-${count.index + 1}"
    Environment = var.name_prefix
  }
}

# Public route table
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = {
    Name        = "${var.name_prefix}-public-rt"
    Environment = var.name_prefix
  }
}

resource "aws_route_table_association" "public" {
  count          = length(var.availability_zones)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Private route tables (one per AZ, each pointing to its NAT gateway)
resource "aws_route_table" "private" {
  count  = var.enable_nat_gateway ? length(var.availability_zones) : 0
  vpc_id = aws_vpc.this.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.this[count.index].id
  }

  tags = {
    Name        = "${var.name_prefix}-private-rt-${count.index + 1}"
    Environment = var.name_prefix
  }
}

resource "aws_route_table_association" "private" {
  count          = var.enable_nat_gateway ? length(var.availability_zones) : 0
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# VPC endpoints (optional, for S3/ECR access without NAT)
resource "aws_vpc_endpoint" "s3" {
  count             = var.enable_nat_gateway ? 0 : 1
  vpc_id            = aws_vpc.this.id
  service_name      = "com.amazonaws.${var.aws_region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = [aws_vpc.this.main_route_table_id]
}

