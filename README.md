# Sistema de Benchmark para Operações S3

Este projeto implementa um sistema para gerar, replicar e fazer upload de arquivos para um bucket S3, além de realizar operações de benchmark como GET, HEAD (STAT) e DELETE, similar à ferramenta Warp do MinIO. O objetivo é avaliar o desempenho de um servidor S3 sob carga.

## Sumário

- [Descrição](#descrição)
- [Características](#características)
- [Pré-requisitos](#pré-requisitos)
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Uso](#uso)
- [Exemplo de Saída](#exemplo-de-saída)
- [Licença](#licença)
- [Contribuições](#contribuições)

## Descrição

O programa realiza as seguintes etapas:

1. **Geração de Arquivos Base**: Cria um conjunto de arquivos base com conteúdo aleatório, se ainda não existirem.
2. **Replicação de Arquivos**: Replica os arquivos base usando o `reflink` para economizar espaço e tempo.
3. **Upload para S3**: Faz upload dos arquivos replicados para um bucket S3 em paralelo.
4. **Benchmarking**: Executa operações GET, HEAD (STAT) e DELETE nos arquivos armazenados no S3 por um período definido.
5. **Relatório Final**: Gera um relatório detalhado com métricas de desempenho das operações realizadas.

## Características

- **Alta Concurrência**: Configuração de nível de concorrência para uploads e operações de benchmark.
- **Operações Paralelas**: Utiliza goroutines e canais para realizar operações em paralelo.
- **Métricas de Desempenho**: Coleta e reporta tempo mínimo, máximo e médio para cada tipo de operação.
- **Configuração Flexível**: Personalize o comportamento do programa através de um arquivo `config.json`.
- **Compatibilidade**: Funciona com qualquer servidor compatível com a API S3.

## Pré-requisitos

- **Go**: Versão 1.16 ou superior instalada.
- **AWS SDK para Go**: Pode ser instalado via `go get`.
- **Credenciais AWS**: Chave de acesso e segredo com permissões para operações S3.
- **Sistema de Arquivos com Suporte a Reflink**: Como Btrfs ou XFS (opcional, mas recomendado).

## Instalação

1. **Clone o Repositório**

   ```bash
   git clone https://github.com/seu-usuario/seu-repositorio.git
   cd seu-repositorio
   ```

2. **Instale as Dependências**

   ```bash
   go get github.com/aws/aws-sdk-go
   ```

3. **Compile o Programa**

   ```bash
   go build -o s3-benchmark
   ```

## Configuração

Crie um arquivo `config.json` na raiz do projeto com o seguinte formato:

```json
{
    "bucketName": "nome-do-seu-bucket",
    "s3Folder": "pasta-base-no-s3",
    "accessKey": "sua-chave-de-acesso",
    "secretKey": "seu-segredo",
    "baseDirectory": "./base_files",
    "minSize": 1024,
    "maxSize": 2048,
    "maxFilesPerFolder": 1000,
    "baseFileCount": 10,
    "totalFiles": 100,
    "maxConcurrentUploads": 10,
    "maxIdleConns": 100,
    "maxIdleConnsPerHost": 100,
    "httpTimeout": 30,
    "maxRetries": 3,
    "endpointURLs": ["https://sua-url-s3"],
    "maxConcurrentReplicas": 5,
    "pauseDurationSeconds": 60,
    "maxLocalFiles": 100,
    "maxBenchmarkThreads": 10,
    "benchmarkDurationSeconds": 60
}
```

**Descrição dos Campos Principais:**

- `bucketName`: Nome do bucket S3 onde os arquivos serão armazenados.
- `s3Folder`: Pasta base dentro do bucket S3.
- `accessKey` e `secretKey`: Credenciais AWS para acesso ao S3.
- `baseDirectory`: Diretório local onde os arquivos base serão gerados.
- `minSize` e `maxSize`: Tamanho mínimo e máximo dos arquivos gerados (em bytes).
- `totalFiles`: Número total de arquivos a serem replicados e enviados.
- `maxConcurrentUploads`: Número máximo de uploads simultâneos.
- `endpointURLs`: Lista de URLs dos endpoints S3.
- `benchmarkDurationSeconds`: Duração (em segundos) das operações de benchmark.

## Uso

1. **Prepare o Arquivo de Configuração**

   Edite o arquivo `config.json` com os parâmetros desejados.

2. **Execute o Programa**

   ```bash
   ./s3-benchmark
   ```

3. **Acompanhe a Saída**

   O programa exibirá informações sobre o progresso das operações, incluindo geração de arquivos, upload e benchmarking.

## Exemplo de Saída

```
Starting pprof server on port 6060
100% completed - 10 base files generated.
Starting file replication with reflink in parallel.
Replicating files: 100/100 (100.00%) | Rate: 5000.00 files/sec
File replication completed.
Replication of 100 files completed.

Uploading to S3: 100/100 (100.00%) | Rate: 2000.00 files/sec
Upload completed for subfolder index 0.
Pausing for 60 seconds before next upload...

All uploads completed.
Performing benchmarking operations...

GET and STAT operations completed. Starting DELETE operations...

Benchmarking Report:
====================

Operation: GET
Total Operations: 5000
Successes: 5000
Errors: 0
Min Time: 5ms
Max Time: 50ms
Avg Time: 25ms

Operation: STAT
Total Operations: 5000
Successes: 5000
Errors: 0
Min Time: 3ms
Max Time: 40ms
Avg Time: 20ms

Operation: DELETE
Total Operations: 100
Successes: 100
Errors: 0
Min Time: 10ms
Max Time: 30ms
Avg Time: 15ms

Overall Benchmark Summary:
Total Operations: 10100
Total Errors: 0
Benchmarking Duration: 1m0s
====================
```

## Licença

Este projeto está licenciado sob a [MIT License](LICENSE).

## Contribuições

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou pull requests no repositório.

---

# S3 Operations Benchmark System

This project implements a system to generate, replicate, and upload files to an S3 bucket, as well as perform benchmarking operations like GET, HEAD (STAT), and DELETE, similar to MinIO's Warp tool. The goal is to assess the performance of an S3 server under load.

## Table of Contents

- [Description](#description)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Sample Output](#sample-output)
- [License](#license)
- [Contributing](#contributing)

## Description

The program performs the following steps:

1. **Base File Generation**: Creates a set of base files with random content if they don't already exist.
2. **File Replication**: Replicates the base files using `reflink` to save space and time.
3. **Upload to S3**: Uploads the replicated files to an S3 bucket in parallel.
4. **Benchmarking**: Performs GET, HEAD (STAT), and DELETE operations on the stored S3 files for a defined period.
5. **Final Report**: Generates a detailed report with performance metrics of the operations performed.

## Features

- **High Concurrency**: Configure concurrency levels for uploads and benchmarking operations.
- **Parallel Operations**: Utilizes goroutines and channels to perform operations in parallel.
- **Performance Metrics**: Collects and reports minimum, maximum, and average times for each operation type.
- **Flexible Configuration**: Customize program behavior via a `config.json` file.
- **Compatibility**: Works with any server compatible with the S3 API.

## Prerequisites

- **Go**: Version 1.16 or higher installed.
- **AWS SDK for Go**: Can be installed via `go get`.
- **AWS Credentials**: Access key and secret with permissions for S3 operations.
- **Filesystem with Reflink Support**: Such as Btrfs or XFS (optional but recommended).

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/your-username/your-repository.git
   cd your-repository
   ```

2. **Install Dependencies**

   ```bash
   go get github.com/aws/aws-sdk-go
   ```

3. **Build the Program**

   ```bash
   go build -o s3-benchmark
   ```

## Configuration

Create a `config.json` file in the project's root directory with the following format:

```json
{
    "bucketName": "your-bucket-name",
    "s3Folder": "your-s3-base-folder",
    "accessKey": "your-access-key",
    "secretKey": "your-secret-key",
    "baseDirectory": "./base_files",
    "minSize": 1024,
    "maxSize": 2048,
    "maxFilesPerFolder": 1000,
    "baseFileCount": 10,
    "totalFiles": 100,
    "maxConcurrentUploads": 10,
    "maxIdleConns": 100,
    "maxIdleConnsPerHost": 100,
    "httpTimeout": 30,
    "maxRetries": 3,
    "endpointURLs": ["https://your-s3-url"],
    "maxConcurrentReplicas": 5,
    "pauseDurationSeconds": 60,
    "maxLocalFiles": 100,
    "maxBenchmarkThreads": 10,
    "benchmarkDurationSeconds": 60
}
```

**Key Field Descriptions:**

- `bucketName`: Name of the S3 bucket where files will be stored.
- `s3Folder`: Base folder within the S3 bucket.
- `accessKey` and `secretKey`: AWS credentials for S3 access.
- `baseDirectory`: Local directory where base files will be generated.
- `minSize` and `maxSize`: Minimum and maximum size of generated files (in bytes).
- `totalFiles`: Total number of files to replicate and upload.
- `maxConcurrentUploads`: Maximum number of simultaneous uploads.
- `endpointURLs`: List of S3 endpoint URLs.
- `benchmarkDurationSeconds`: Duration (in seconds) for benchmarking operations.

## Usage

1. **Prepare the Configuration File**

   Edit the `config.json` file with your desired parameters.

2. **Run the Program**

   ```bash
   ./s3-benchmark
   ```

3. **Monitor the Output**

   The program will display information about the progress of operations, including file generation, upload, and benchmarking.

## Sample Output

```
Starting pprof server on port 6060
100% completed - 10 base files generated.
Starting file replication with reflink in parallel.
Replicating files: 100/100 (100.00%) | Rate: 5000.00 files/sec
File replication completed.
Replication of 100 files completed.

Uploading to S3: 100/100 (100.00%) | Rate: 2000.00 files/sec
Upload completed for subfolder index 0.
Pausing for 60 seconds before next upload...

All uploads completed.
Performing benchmarking operations...

GET and STAT operations completed. Starting DELETE operations...

Benchmarking Report:
====================

Operation: GET
Total Operations: 5000
Successes: 5000
Errors: 0
Min Time: 5ms
Max Time: 50ms
Avg Time: 25ms

Operation: STAT
Total Operations: 5000
Successes: 5000
Errors: 0
Min Time: 3ms
Max Time: 40ms
Avg Time: 20ms

Operation: DELETE
Total Operations: 100
Successes: 100
Errors: 0
Min Time: 10ms
Max Time: 30ms
Avg Time: 15ms

Overall Benchmark Summary:
Total Operations: 10100
Total Errors: 0
Benchmarking Duration: 1m0s
====================
```

## License

This project is licensed under the [MIT License](LICENSE).

## Contributing

Contributions are welcome! Feel free to open issues or pull requests in the repository.

---

**Note**: Make sure to update the `LICENSE` file with the appropriate license for your project.
