package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"io/ioutil"
	"regexp"
)

// LangConfig représente la configuration pour chaque langage
type LangConfig struct {
	DockerImage string `json:"docker_image"`
	Command     string `json:"command"`
}

// ExecuteCode exécute le code dans un conteneur Docker
func ExecuteCode(lang string, file string) {
	// Charger la configuration
	langConfig := loadLangConfig(lang)
	if langConfig == nil {
		fmt.Printf("Langage non supporté : %s\n", lang)
		return
	}

	// Copier le fichier dans le dossier temporaire
	tempDir := "/tmp/auto-tester-cli"
	os.MkdirAll(tempDir, os.ModePerm)
	dstFile := filepath.Join(tempDir, filepath.Base(file))
	copyFile(file, dstFile)

	// Générer un fichier de test
	testFile := generateTestFile(file)

	// Copier le fichier de test dans le dossier temporaire
	dstTestFile := filepath.Join(tempDir, "test_"+filepath.Base(file))
	copyFile(testFile, dstTestFile)

	// Construire et exécuter le conteneur Docker
	runDockerContainer(langConfig.DockerImage, dstTestFile, langConfig.Command)
}

func loadLangConfig(lang string) *LangConfig {
	configFile := "./configs/languages.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("Erreur de lecture du fichier de configuration :", err)
		return nil
	}

	// Décoder la configuration JSON
	var configs map[string]LangConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		fmt.Println("Erreur de parsing JSON :", err)
		return nil
	}

	// Retourner la configuration spécifique au langage
	config, exists := configs[lang]
	if !exists {
		return nil
	}
	return &config
}

func copyFile(src, dst string) {
	input, err := os.ReadFile(src)
	if err != nil {
		fmt.Println("Erreur de lecture du fichier :", err)
		os.Exit(1)
	}
	if err := os.WriteFile(dst, input, 0644); err != nil {
		fmt.Println("Erreur d'écriture du fichier :", err)
		os.Exit(1)
	}
}

func runDockerContainer(image, file, command string) {
    fmt.Println("Exécution du code avec l'image :", image)

    // Obtenir le nom du fichier
    fileName := filepath.Base(file)

    // Ajuster la commande pour exécuter le fichier
    fullCommand := fmt.Sprintf(command, fileName)

    // Créer une commande Docker pour exécuter le conteneur
    cmd := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/app/%s", file, fileName), image, "/bin/sh", "-c", fullCommand)

    // Capturer la sortie du conteneur
    out, err := cmd.CombinedOutput()

    // Afficher la sortie complète (logs)
    fmt.Println(string(out))

    // Vérifier si une erreur s'est produite et ajuster le code en fonction
    if err != nil {
        fmt.Printf("Erreur lors de l'exécution du conteneur : %s\n", err)
        os.Exit(1)
    }

    // Ici, tu peux analyser les logs de sortie et calculer les ratios
    analyzeLogs(string(out))
}

// Génère un fichier de tests à partir du code source
// Génère un fichier de tests à partir du code source
func generateTestFile(file string) string {
	// Lire le contenu du fichier source
	content, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier :", err)
		os.Exit(1)
	}

	// Utiliser une expression régulière pour extraire les fonctions Python
	re := regexp.MustCompile(`def (\w+)\((.*?)\):`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	// Générer un fichier de test simple basé sur les fonctions trouvées
	var testContent string
	for _, match := range matches {
		functionName := match[1]
		params := match[2]

		// Initialisation des paramètres (ici, on suppose que les fonctions ont 2 paramètres)
		paramsList := strings.Split(params, ",")
		testParams := make([]string, len(paramsList))
		for i := range paramsList {
			testParams[i] = fmt.Sprintf("param%d", i+1) // Utilisation de param1, param2, ...
		}

		// Générer un test pour chaque fonction avec des valeurs fixes
		testContent += fmt.Sprintf(`
def test_%s():
    %s = %d  # Exemple de valeur
    %s = %d  # Exemple de valeur
    assert %s(%s) == %d  # Attendu : résultat de l'addition de %s et %s
`, functionName, testParams[0], 2, testParams[1], 3, functionName, strings.Join(testParams, ","), 5, testParams[0], testParams[1])
	}

	// Sauvegarder le fichier de test généré
	testFile := "./generated_tests.py"
	if err := ioutil.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		fmt.Println("Erreur d'écriture du fichier de test :", err)
		os.Exit(1)
	}

	return testFile
}


// Fonction pour analyser les logs et calculer les ratios
func analyzeLogs(logs string) {
    var totalTests, passedTests, failedTests int
    var successRatio float64

    // Ici, on imagine que les logs contiennent des lignes avec "Test Passed" et "Test Failed"
    // Cela peut être ajusté selon le format exact des logs de ton application
    lines := strings.Split(logs, "\n")
    for _, line := range lines {
        if strings.Contains(line, "Test Passed") {
            passedTests++
        } else if strings.Contains(line, "Test Failed") {
            failedTests++
        }
        totalTests++
    }

    // Calcul du ratio de succès
    if totalTests > 0 {
        successRatio = float64(passedTests) / float64(totalTests) * 100
    }

    // Affichage des résultats détaillés
    fmt.Printf("\n=== Résumé des tests ===\n")
    fmt.Printf("Total des tests exécutés : %d\n", totalTests)
    fmt.Printf("Tests réussis : %d\n", passedTests)
    fmt.Printf("Tests échoués : %d\n", failedTests)
    fmt.Printf("Ratio de succès : %.2f%%\n", successRatio)
}

