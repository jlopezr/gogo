package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	fmt.Println("Iniciando aplicacion...")
	
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)		
	}
	client := s3.NewFromConfig(cfg)

	err = ListarCarpetaS3(client, "mi-bucket", "mi/prefijo/")
	if err != nil {
		log.Fatalf("Error al listar: %v", err)
	}

}
