package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3File struct {
	Key          string
	LastModified time.Time
}

func main() {
	fmt.Println("Iniciando aplicacion...")

	fmt.Println("Plugins registrados:")
	for _, name := range ListPlugins() {
		fmt.Println("-", name)
	}

	plugin, found := GetPlugin("plugin1")
	if found {
		fmt.Println(plugin.Execute())
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}
	client := s3.NewFromConfig(cfg)

	err = ListarCarpetaS3(client, "mi-bucket", "mi/prefijo/")
	if err != nil {
		log.Fatalf("Error al listar: %v", err)
	}

	directorios := map[string]string{
		"directorio1/": "proceso1",
		"directorio2/": "proceso2",
		"directorio3/": "proceso3",
	}

	var wg sync.WaitGroup

	for dir, proceso := range directorios {
		wg.Add(1)
		go func(dir, proceso string) {
			defer wg.Done()
			err := procesarDirectorio(client, "mi-bucket", dir, proceso)
			if err != nil {
				log.Printf("Error procesando %s: %v", dir, err)

			}
		}(dir, proceso)
	}

	wg.Wait()
	fmt.Println("Todos los directorios han sido procesados")
}

func procesarDirectorio(client *s3.Client, bucket, prefix, proceso string) error {
	archivos, err := listarArchivos(client, bucket, prefix)
	if err != nil {
		return err
	}

	for _, archivo := range archivos {
		fmt.Println("Procesando %s en %s con %s...\n", archivo.Key, prefix, proceso)
		if err := ejecutarProcesoExterno(proceso, archivo.Key); err != nil {
			log.Printf("Error procesando archivo %s: %v", archivo.Key, err)
			continue
		}
		log.Printf("Archivo %s procesado.\n", archivo.Key)
	}

	return nil
}

func ejecutarProcesoExterno(proceso, archivo string) error {
	fmt.Printf("Ejecutando %s para %s...\n", proceso, archivo)
	time.Sleep(2 * time.Second)
	fmt.Printf("Proceso %s completado para %s.\n", proceso, archivo)
	return nil
}
