package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

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

	client := ConfigurarClienteS3("", "us-east-1", "ACCESSKEY", "SECRETKEY")

	err := ListarCarpetaS3(client, "mi-bucket", "mi/prefijo/")
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
	archivos, err := listarArchivosConPaginacion(client, bucket, prefix)
	if err != nil {
		return err
	}

	for _, archivo := range archivos {
		fmt.Printf("Procesando %s en %s con %s...\n", archivo.Key, prefix, proceso)
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

// Estado y comandos

type EstadoDirectorio struct {
	Actual     string   `json:"actual"`     // Archivo que se esta procesando
	Pendientes []string `json:"pendientes"` // Archivos pendientes
}

var estado = struct {
	sync.Mutex
	Directorio map[string]*EstadoDirectorio
}{
	Directorio: make(map[string]*EstadoDirectorio),
}

func ejecutarRun(directorios map[string]string) {
	// Asegurarnos que solo un programa se ejecute
	lockFile := "/tmp/program.lock"
	lock, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		fmt.Println("Otro proceso ya se esta ejecutando")
		return
	}
	defer os.Remove(lockFile)
	defer lock.Close()

	// Iniciar servidor para status
	go iniciarServidorStatus()

	var wg sync.WaitGroup
	for dir, proceso := range directorios {
		wg.Add(1)
		go func(dir, proceso string) {
			defer wg.Done()
			procesarDirectorio2(dir, proceso)
		}(dir, proceso)
	}

	wg.Wait()
	fmt.Println("Procesamiento completado")
}

func ejecutarStatus() {
	conn, err := net.Dial("unix", "/tmp/program.sock")
	if err != nil {
		fmt.Println("No se pudo conectar al proceso en ejecución")
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("status"))
	if err != nil {
		log.Fatalf("Error enviando comando status; %v", err)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Error leyendo respuesta: %v", err)
	}

	fmt.Println("Estado actual:")
	fmt.Println(string(buf[:n]))
}

func iniciarServidorStatus() {
	listener, err := net.Listen("unix", "/tmp/program.sock")
	if err != nil {
		log.Fatalf("Error iniciando servidor de status: %v", err)
	}
	defer os.Remove("/tmp/program.sock")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error aceptando conexión: %v", err)
			continue
		}
		go manejarConexion(conn)
	}
}

func manejarConexion(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Error leyendo comando: %v", err)
		return
	}

	comando := string(buf[:n])
	if comando == "status" {
		estado.Lock()
		defer estado.Unlock()

		respuesta, err := json.Marshal(estado.Directorio)
		if err != nil {
			log.Printf("Error serializando estado: %v", err)
			return
		}

		_, err = conn.Write(respuesta)
		if err != nil {
			log.Printf("Error enviando respuesta: %v", err)
		}

	}
}

// TODO Esto esta simulado
func procesarDirectorio2(dir, proceso string) {
	archivos := []string{"archivo1.txt", "archivo2.txt", "archivo3.txt"}

	estado.Lock()
	estado.Directorio[dir] = &EstadoDirectorio{
		Actual:     "",
		Pendientes: archivos,
	}
	estado.Unlock()

	for len(archivos) > 0 {
		estado.Lock()
		actual := archivos[0]
		archivos = archivos[1:]
		estado.Directorio[dir].Actual = actual
		estado.Directorio[dir].Pendientes = archivos
		estado.Unlock()

		fmt.Printf("Procesando %s en %s con %s...", actual, dir, proceso)
		time.Sleep(2 * time.Second)
		fmt.Printf("Procesado %s\n", actual)
	}

	estado.Lock()
	delete(estado.Directorio, dir)
	estado.Unlock()
}
