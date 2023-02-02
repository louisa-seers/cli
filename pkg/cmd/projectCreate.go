/*
 * Copyright contributors to the Galasa project
 */
package cmd

import (
	"log"
	"strings"

	"github.com/galasa.dev/cli/pkg/embedded"
	galasaErrors "github.com/galasa.dev/cli/pkg/errors"

	"github.com/galasa.dev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

// CommonPomParameters holds common substitution parameters a pom.xml file
// template uses.
type MavenCoordinates struct {
	GroupId    string
	ArtifactId string
	Name       string
}

var (
	projectCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates a new Galasa project",
		Long:  "Creates a new Galasa test project with optional OBR project and build process files",
		Args:  cobra.NoArgs,
		Run:   executeCreateProject,
	}

	packageName                string
	force                      bool
	isOBRProjectRequired       bool
	featureNamesCommaSeparated string
)

func init() {
	cmd := projectCreateCmd
	parentCommand := projectCmd

	cmd.Flags().StringVar(&packageName, "package", "", "Java package name for tests we create. "+
		"Forms part of the project name, maven/gradle group/artifact ID, "+
		"and OSGi bundle name. It may reflect the name of your organisation or company, "+
		"the department, function or application under test. "+
		"For example: dev.galasa.banking.example")
	cmd.MarkFlagRequired("package")

	cmd.Flags().BoolVar(&force, "force", false, "Force-overwrite files which already exist.")
	cmd.Flags().BoolVar(&isOBRProjectRequired, "obr", false, "An OSGi Object Bundle Resource (OBR) project is needed.")
	cmd.Flags().StringVar(&featureNamesCommaSeparated, "features", "main",
		"A comma-separated list of features you are testing. Defaults to \"test\". "+
			"These must be able to form parts of a java package name. "+
			"For example: \"payee,account\"")
	parentCommand.AddCommand(cmd)
}

func executeCreateProject(cmd *cobra.Command, args []string) {

	utils.CaptureLog(logFileName)

	log.Println("Galasa CLI - Create project")

	// Operations on the file system will all be relative to the current folder.
	fileSystem := utils.NewOSFileSystem()

	err := createProject(fileSystem, packageName, featureNamesCommaSeparated, isOBRProjectRequired, force)

	// Convey the error to the top level.
	// Tried doing this with RunE: entry, passing back the error, but we always
	// got a 'usage' syntax summary for the command which failed.
	if err != nil {
		// We can't unit test
		panic(err)
	}
}

// createProject will create the following artifacts in the specified file system:
//
//		. - All files are relative to the current directory.
//		└── packageName - The parent package
//			├── pom.xml - The parent pom.
//	 		├── packageName.test - The tests project.
//	  		│   └── pom.xml
//	  		│   └── src
//	  		│       └── main
//	  		│           └── java
//	  		│               └── packageName - There will be multiple nested folders if there are dots ('.') in the package name
//	  		│                   └── SampleTest.java - Contains an example Galasa testcase
//	  		└── packageName.obr - The OBR project. (only if the --obr option is used).
//	    	 	 └── pom.xml
//
// isOBRProjectRequired - Controls whether the optional OBR project is going to be created.
// featureNamesCommaSeparated - eg: kettle,toaster. Causes a kettle and toaster project to be created with a sample test in.
func createProject(
	fileSystem utils.FileSystem,
	packageName string,
	featureNamesCommaSeparated string,
	isOBRProjectRequired bool,
	forceOverwrite bool) error {

	log.Printf("Creating project using packageName:%s\n", packageName)

	var err error

	embeddedFileSystem := embedded.GetEmbeddedFileSystem()

	fileGenerator := utils.NewFileGenerator(fileSystem, embeddedFileSystem)

	// Separate out the feature names from a string into a slice of strings.
	var featureNames []string
	featureNames, err = separateFeatureNamesFromCommaSeparatedList(featureNamesCommaSeparated)

	if err == nil {
		err = utils.ValidateJavaPackageName(packageName)
		if err == nil {
			// Create the parent folder
			parentProjectFolder := packageName
			err = fileGenerator.CreateFolder(parentProjectFolder)
			if err == nil {
				err = createParentFolderPom(fileGenerator, packageName, featureNames, isOBRProjectRequired, forceOverwrite)
				if err == nil {
					err = createTestProjects(fileGenerator, packageName, featureNames, forceOverwrite)
					if err == nil {
						if isOBRProjectRequired {
							err = createOBRProject(fileGenerator, packageName, featureNames, forceOverwrite)
						}
					}
				}
			}
		}
	}

	return err
}

func separateFeatureNamesFromCommaSeparatedList(featureNamesCommaSeparated string) ([]string, error) {
	featureNames := strings.Split(featureNamesCommaSeparated, ",")

	var err error

	// Validate each feature name can form part of a package...
	// meaning it should be a valid package name in it's own right.
	for _, featureName := range featureNames {
		err = utils.ValidateJavaPackageName(featureName)
		if err != nil {
			err = galasaErrors.NewGalasaError(galasaErrors.GALASA_ERROR_INVALID_FEATURE_NAME, featureName, err.Error())
			featureNames = nil
			break
		}
	}

	return featureNames, err
}

func createParentFolderPom(fileGenerator *utils.FileGenerator, packageName string, featureNames []string, isOBRRequired bool, forceOverwrite bool) error {

	type ParentPomParameters struct {
		Coordinates MavenCoordinates

		// Version of Galasa we are targetting
		GalasaVersion string

		IsOBRRequired    bool
		ObrName          string
		ChildModuleNames []string
	}

	templateParameters := ParentPomParameters{
		Coordinates:      MavenCoordinates{ArtifactId: packageName, GroupId: packageName, Name: packageName},
		GalasaVersion:    embedded.GetGalasaVersion(),
		IsOBRRequired:    isOBRRequired,
		ObrName:          packageName + ".obr",
		ChildModuleNames: make([]string, len(featureNames))}
	// Populate the child module names
	for index, featureName := range featureNames {
		templateParameters.ChildModuleNames[index] = packageName + "." + featureName
	}

	targetFile := utils.GeneratedFileDef{
		FileType:                 "pom",
		TargetFilePath:           packageName + "/pom.xml",
		EmbeddedTemplateFilePath: "templates/projectCreate/parent-project/pom.xml",
		TemplateParameters:       templateParameters}

	err := fileGenerator.CreateFile(targetFile, forceOverwrite, true)
	return err
}

// createTestProjects - creates a number of projects, each of whichh containing tests which test a feature.
func createTestProjects(fileGenerator *utils.FileGenerator, packageName string, featureNames []string, forceOverwrite bool) error {
	var err error = nil
	for _, featureName := range featureNames {
		err = createTestProject(fileGenerator, packageName, featureName, forceOverwrite)
		if err != nil {
			break
		}
	}
	return err
}

// createTestProject - creates a single project to contain tests which test a feature.
func createTestProject(
	fileGenerator *utils.FileGenerator,
	packageName string,
	featureName string,
	forceOverwrite bool) error {

	targetFolderPath := packageName + "/" + packageName + "." + featureName
	log.Printf("Creating tests project %s\n", targetFolderPath)

	// Create the base test folder
	err := fileGenerator.CreateFolder(targetFolderPath)
	if err == nil {
		err = createTestFolderPom(fileGenerator, targetFolderPath, packageName, featureName, forceOverwrite)
	}

	if err == nil {
		err = createJavaSourceFolder(fileGenerator, targetFolderPath, packageName, featureName, forceOverwrite)
	}

	if err == nil {
		err = createTestResourceFolder(fileGenerator, targetFolderPath, packageName, forceOverwrite)
	}

	if err == nil {
		log.Printf("Tests project %s created OK.", targetFolderPath)
	}
	return err
}

func createOBRProject(fileGenerator *utils.FileGenerator, packageName string, featureNames []string, forceOverwrite bool) error {
	targetFolderPath := packageName + "/" + packageName + ".obr"
	log.Printf("Creating obr project %s\n", targetFolderPath)

	// Create the base test folder
	err := fileGenerator.CreateFolder(targetFolderPath)
	if err == nil {
		err = createOBRFolderPom(fileGenerator, targetFolderPath, packageName, featureNames, forceOverwrite)
	}

	if err == nil {
		log.Printf("OBR project %s created OK.", targetFolderPath)
	}
	return err
}

func createJavaSourceFolder(fileGenerator *utils.FileGenerator, testFolderPath string, packageName string, featureName string, forceOverwrite bool) error {

	// The folder is the package name but with slashes.
	// eg: my.package becomes my/package
	packageNameWithSlashes := strings.Replace(packageName, ".", "/", -1)
	targetSrcFolderPath := testFolderPath + "/src/main/java/" + packageNameWithSlashes + "/" + featureName
	err := fileGenerator.CreateFolder(targetSrcFolderPath)
	if err == nil {
		// Create our first test java source file.
		classNameNoClassSuffix := "Test" + utils.UppercaseFirstLetter(featureName)
		templateBundlePath := "templates/projectCreate/parent-project/test-project/src/main/java/TestSimple.java.template"
		err = createJavaSourceFile(fileGenerator, targetSrcFolderPath, packageName,
			featureName, forceOverwrite, classNameNoClassSuffix, templateBundlePath)

		if err == nil {
			// Create our second test java source file. To show that you can have multiples.
			classNameNoClassSuffix = "Test" + utils.UppercaseFirstLetter(featureName) + "Extended"
			templateBundlePath := "templates/projectCreate/parent-project/test-project/src/main/java/TestExtended.java.template"
			err = createJavaSourceFile(fileGenerator, targetSrcFolderPath, packageName,
				featureName, forceOverwrite, classNameNoClassSuffix, templateBundlePath)
		}
	}
	return err
}

func createTestResourceFolder(
	fileGenerator *utils.FileGenerator, targetSrcFolderPath string,
	packageName string, forceOverwrite bool) error {

	var err error

	// Create the folder for the resources to sit in.
	targetResourceFolderPath := targetSrcFolderPath + "/src/main/resources/textfiles"
	err = fileGenerator.CreateFolder(targetResourceFolderPath)
	if err == nil {

		targetFilePath := targetResourceFolderPath + "/sampleText.txt"
		templateBundlePath := "templates/projectCreate/parent-project/test-project/src/main/resources/textfiles/sampleText.txt"

		targetFile := utils.GeneratedFileDef{
			FileType:                 "TextFile",
			TargetFilePath:           targetFilePath,
			EmbeddedTemplateFilePath: templateBundlePath,
			TemplateParameters:       nil}

		err = fileGenerator.CreateFile(targetFile, forceOverwrite, true)
	}
	return err
}

func createJavaSourceFile(fileGenerator *utils.FileGenerator, targetSrcFolderPath string,
	packageName string, featureName string, forceOverwrite bool,
	classNameNoClassSuffix string, templateBundlePath string) error {

	// JavaTestTemplateSubstitutionParameters holds all the substitution parameters a java test file
	// template uses
	type JavaTestTemplateSubstitutionParameters struct {
		Package   string
		ClassName string
	}

	templateParameters := JavaTestTemplateSubstitutionParameters{
		Package:   packageName + "." + featureName,
		ClassName: classNameNoClassSuffix}

	targetFile := utils.GeneratedFileDef{
		FileType:                 "JavaSourceFile",
		TargetFilePath:           targetSrcFolderPath + "/" + classNameNoClassSuffix + ".java",
		EmbeddedTemplateFilePath: templateBundlePath,
		TemplateParameters:       templateParameters}

	err := fileGenerator.CreateFile(targetFile, forceOverwrite, true)
	return err
}

func createTestFolderPom(fileGenerator *utils.FileGenerator, targetTestFolderPath string,
	packageName string, featureName string, forceOverwrite bool) error {

	type TestPomParameters struct {
		Parent      MavenCoordinates
		Coordinates MavenCoordinates
		FeatureName string
	}

	pomTemplateParameters := TestPomParameters{
		Parent:      MavenCoordinates{GroupId: packageName, ArtifactId: packageName, Name: packageName},
		Coordinates: MavenCoordinates{GroupId: packageName, ArtifactId: packageName + "." + featureName, Name: packageName + "." + featureName},
		FeatureName: featureName}

	targetFile := utils.GeneratedFileDef{
		FileType:                 "pom",
		TargetFilePath:           targetTestFolderPath + "/pom.xml",
		EmbeddedTemplateFilePath: "templates/projectCreate/parent-project/test-project/pom.xml",
		TemplateParameters:       pomTemplateParameters}

	err := fileGenerator.CreateFile(targetFile, forceOverwrite, true)
	return err
}

func createOBRFolderPom(fileGenerator *utils.FileGenerator, targetOBRFolderPath string, packageName string,
	featureNames []string, forceOverwrite bool) error {

	type OBRPomParameters struct {
		Parent      MavenCoordinates
		Coordinates MavenCoordinates
		Modules     []MavenCoordinates
	}

	// Fill-in all the parameters the template needs.
	pomTemplateParameters := OBRPomParameters{
		Parent:      MavenCoordinates{GroupId: packageName, ArtifactId: packageName, Name: packageName},
		Coordinates: MavenCoordinates{GroupId: packageName, ArtifactId: packageName + ".obr", Name: packageName + ".obr"},
		Modules:     make([]MavenCoordinates, len(featureNames))}
	// Populate the list of modules.
	for index, featureName := range featureNames {
		pomTemplateParameters.Modules[index] = MavenCoordinates{
			GroupId: packageName, ArtifactId: packageName + "." + featureName, Name: packageName + "." + featureName}
	}

	targetFile := utils.GeneratedFileDef{
		FileType:                 "pom",
		TargetFilePath:           targetOBRFolderPath + "/pom.xml",
		EmbeddedTemplateFilePath: "templates/projectCreate/parent-project/obr-project/pom.xml",
		TemplateParameters:       pomTemplateParameters}

	err := fileGenerator.CreateFile(targetFile, forceOverwrite, true)
	return err
}