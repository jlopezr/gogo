package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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
		Key: aws.String(key),
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
		Key: aws.String(key),
		Body: buffer,
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		return err
	}

	fmt.Printf("Archivo '%s' subido como '%s' al bucket '%s'\n", archivo, key, bucket)
	return nil
}
