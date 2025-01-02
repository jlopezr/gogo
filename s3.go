package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ConfigurarClienteS3(endpoint, region, accessKey, secretKey string) *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}
	client := s3.NewFromConfig(cfg)

	return client
}

func ListarCarpetaS3(client *s3.Client, bucket, prefix string) error {
	fmt.Println("Funcion en archivo s3.go")

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	resp, err := client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return err
	}

	fmt.Printf("Objectos en el bucket '%s' con prefijo '%s':\n", bucket, prefix)
	for _, obj := range resp.Contents {
		fmt.Printf("- %s (tamaño: %d bytes)", *obj.Key, obj.Size)
	}
	return nil
}

func DescargarObjetoS3(client *s3.Client, bucket, key, destino string) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	resp, err := client.GetObject(context.TODO(), input)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Objeto '%s' descargado a %s\n", key, destino)
	return nil
}

func SubirObjetoS3(client *s3.Client, bucket, key, archivo string) error {
	file, err := os.Open(archivo)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(file)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   buffer,
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		return err
	}

	fmt.Printf("Archivo '%s' subido como '%s' al bucket '%s'\n", archivo, key, bucket)
	return nil
}

func CambiarNombreObjetoS3(client *s3.Client, bucket, key string) error {
	nuevoNombre := "*" + key //TODO Poner la * en el ultimo elemento del path

	_, err := client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		CopySource: aws.String(bucket + "/" + key),
		Key:        aws.String(nuevoNombre),
	})
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Objeto '%s' renombrado a '%s' en el bucket '%s'\n", key, nuevoNombre, bucket)
	return nil
}

func listarArchivos(client *s3.Client, bucket, prefix string) ([]S3File, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	resp, err := client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var archivos []S3File
	for _, obj := range resp.Contents {
		// Filtrar archivos que no empiecen por '*'
		if !strings.HasPrefix(*obj.Key, "*") {
			archivos = append(archivos, S3File{
				Key:          *obj.Key,
				LastModified: *obj.LastModified,
			})
		}
	}

	// Ordenar por fecha de modificacion
	sort.Slice(archivos, func(i, j int) bool {
		return archivos[i].LastModified.Before(archivos[j].LastModified)
	})

	return archivos, nil
}

func listarArchivosConPaginacion(client *s3.Client, bucket, prefix string) ([]S3File, error) {
	var archivos []S3File
	var continuationToken *string

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		resp, err := client.ListObjectsV2(context.TODO(), input)
		if err != nil {
			return nil, fmt.Errorf("error al listar objetos: %w", err)
		}

		for _, obj := range resp.Contents {
			// Filtrar archivos que no empiecen por '*'
			if !strings.HasPrefix(*obj.Key, "*") {
				archivos = append(archivos, S3File{
					Key:          *obj.Key,
					LastModified: *obj.LastModified,
				})
			}
		}

		// Si no hay mas paginas salir del bucle
		if !*resp.IsTruncated {
			break
		}

		continuationToken = resp.NextContinuationToken
	}

	// Ordenar por fecha de modificacion
	sort.Slice(archivos, func(i, j int) bool {
		return archivos[i].LastModified.Before(archivos[j].LastModified)
	})

	return archivos, nil
}
