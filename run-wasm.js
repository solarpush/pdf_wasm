#!/usr/bin/env node

// Module pour utiliser le WASM PDF Template dans une application Node.js

const fs = require("fs");
const path = require("path");
const { WASI } = require("wasi");
const { Readable, Writable } = require("stream");

// Fonction utilitaire pour s'assurer qu'un dossier existe
function ensureDirectoryExists(filePath) {
  const dir = path.dirname(filePath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

const WASM_PATH = "./pdf-template.wasm";
const WASM_GZ_PATH = "./pdf-template.wasm.gz";
class PDFWasmGenerator {
  constructor(options = {}) {
    this.wasmInstance = null;
    this.wasmModule = null;
    this.preferCompressed = options.preferCompressed ?? true; // Par défaut: compressé
  }

  // Initialiser le module WASM (à faire une seule fois)
  async initialize() {
    if (this.wasmInstance) return; // Déjà initialisé

    const startTime = performance.now();
    let wasmBuffer;
    let loadMethod;

    // Choix selon la préférence utilisateur
    const useCompressed = this.preferCompressed && fs.existsSync(WASM_GZ_PATH);

    if (useCompressed) {
      const loadStart = performance.now();
      console.log("📦 Chargement de la version compressée (1.6MB)...");
      const zlib = require("zlib");
      const compressedBuffer = fs.readFileSync(WASM_GZ_PATH);
      const loadTime = performance.now() - loadStart;

      const decompressStart = performance.now();
      wasmBuffer = zlib.gunzipSync(compressedBuffer);
      const decompressTime = performance.now() - decompressStart;

      loadMethod = `compressée (lecture: ${loadTime.toFixed(
        1
      )}ms, décompression: ${decompressTime.toFixed(1)}ms)`;
    }
    // Fallback vers version normale
    else if (fs.existsSync(WASM_PATH)) {
      const loadStart = performance.now();
      console.log("📦 Chargement de la version standard (5.7MB)...");
      wasmBuffer = fs.readFileSync(WASM_PATH);
      const loadTime = performance.now() - loadStart;

      loadMethod = `standard (lecture: ${loadTime.toFixed(1)}ms)`;
    }
    // Erreur: aucun fichier trouvé
    else {
      throw new Error("❌ Aucun fichier WASM trouvé. Exécutez: make build");
    }

    const compileStart = performance.now();
    this.wasmModule = await WebAssembly.compile(wasmBuffer);
    const compileTime = performance.now() - compileStart;

    const totalTime = performance.now() - startTime;
    console.log(
      `⚡ Initialisation ${loadMethod}, compilation: ${compileTime.toFixed(
        1
      )}ms, total: ${totalTime.toFixed(1)}ms`
    );
  }

  // Générer un PDF à partir d'un template et de variables
  async generatePDF(template, variables) {
    const loadStart = performance.now();
    if (!this.wasmModule) {
      throw new Error(
        "Module WASM non initialisé. Appelez initialize() d'abord."
      );
    }

    return new Promise((resolve, reject) => {
      try {
        // Créer le JSON combiné
        const jsonData = {
          pdf_template:
            typeof template === "string" ? JSON.parse(template) : template,
          pdfVars: variables,
        };

        const inputData = JSON.stringify(jsonData);

        // Configuration WASI simplifiée - utilisation de fichiers temporaires
        const os = require("os");
        const path = require("path");

        const tempDir = os.tmpdir();
        const inputFile = path.join(tempDir, `pdf-input-${Date.now()}.json`);
        const outputFile = path.join(tempDir, `pdf-output-${Date.now()}.pdf`);

        // Écrire le JSON dans un fichier temporaire
        fs.writeFileSync(inputFile, inputData);

        // Configuration WASI avec redirection des fichiers
        const inputFd = fs.openSync(inputFile, "r");
        const outputFd = fs.openSync(outputFile, "w");

        const wasi = new WASI({
          version: "preview1",
          args: ["pdf-template"], // argv[0] = nom du programme
          env: process.env,
          stdin: inputFd,
          stdout: outputFd,
          stderr: 2, // stderr vers la console
          preopens: {
            "/": process.cwd(),
            "/tmp": tempDir,
            ".": process.cwd(), // Accès explicite au répertoire courant
            "/fonts": process.cwd(), // Alias pour les polices
          },
        });

        // Créer l'instance WASM avec la nouvelle API
        WebAssembly.instantiate(this.wasmModule, wasi.getImportObject())
          .then((instance) => {
            try {
              // Démarrer le programme WASM
              wasi.start(instance);

              // Fermer les descripteurs de fichiers
              fs.closeSync(inputFd);
              fs.closeSync(outputFd);

              // Lire le fichier de sortie généré
              const pdfBuffer = fs.readFileSync(outputFile);

              // Nettoyer les fichiers temporaires
              try {
                fs.unlinkSync(inputFile);
                fs.unlinkSync(outputFile);
              } catch (cleanupError) {
                console.warn("Warning: cleanup error:", cleanupError.message);
              }

              resolve(pdfBuffer);
            } catch (wasmError) {
              // Fermer les descripteurs en cas d'erreur
              try {
                fs.closeSync(inputFd);
                fs.closeSync(outputFd);
              } catch (closeError) {
                // Ignorer les erreurs de fermeture
              }
              reject(wasmError);
            }
          })
          .catch(reject);
      } catch (error) {
        reject(error);
      }
      const loadTime = performance.now() - loadStart;
      console.log(`⚡ Génération du pdf, temps: ${loadTime.toFixed(2)}ms`);
    });
  }
}

// Fonction helper pour usage simple
async function generatePDFFromTemplate(templatePath, variables) {
  const generator = new PDFWasmGenerator();
  await generator.initialize();

  const template = fs.readFileSync(templatePath, "utf-8");
  return await generator.generatePDF(template, variables);
}

// Exemple d'utilisation dans ton serveur
async function mainProcess() {
  const values = {
    company: {
      name: "DIGITAL SOLUTIONS SARL",
      address: "123 Avenue des Champs-Élysées",
      city: "75008 Paris, France",
      phone: "+33 1 42 56 78 90",
      email: "contact@digitalsolutions.fr",
      siret: "123 456 789 00012",
      vat: "FR12345678901",
    },
    invoice: {
      number: "FAC-2025-0156",
      date: "31 août 2025",
      dueDate: "30 septembre 2025",
    },
    client: {
      name: "TECH CORP SAS",
      address: "456 Rue de la Innovation",
      city: "69000 Lyon, France",
      country: "France",
    },
    delivery: {
      name: "TECH CORP SAS - Siège",
      address: "456 Rue de la Innovation",
      city: "69000 Lyon, France",
      country: "France",
    },
    items: [
      {
        description: "Développement application web responsive",
        quantity: "1",
        unitPrice: "2800.00",
        vat: "20",
        total: "2800.00",
      },
      {
        description: "Formation équipe (3 jours)",
        quantity: "3",
        unitPrice: "650.00",
        vat: "20",
        total: "1950.00",
      },
    ],
    currency: "€",
    totals: {
      subtotal: "7000.00",
      shipping: "50.00",
      discount: "-200.00",
      ht: "6850.00",
      vatRate: "20",
      vat: "1370.00",
      ttc: "8220.00",
    },
    payment: {
      terms: "Paiement à 30 jours fin de mois",
      method: "Virement bancaire ou chèque",
      iban: "FR14 2004 1010 0505 0001 3M02 606",
      bic: "PSSTFRPPPAR",
    },
    theme: {
      primary: "#1E40AF",
      secondary: "#059669",
    },
    notes:
      "Merci pour votre confiance. En cas de retard de paiement, des pénalités de 3 fois le taux légal seront appliquées.\nTout litige relève de la compétence du Tribunal de Commerce de Paris.",
  };

  try {
    // Méthode 1: Avec la classe (recommandée pour plusieurs PDFs)
    const generator = new PDFWasmGenerator();
    await generator.initialize();

    const template = JSON.parse(fs.readFileSync("template_test.json", "utf-8"));
    const pdfBuffer = await generator.generatePDF(
      template.pdf_template,
      values
    );

    // Sauvegarder ou retourner le buffer
    const outputPath = "output/generated.pdf";
    ensureDirectoryExists(outputPath); // Créer tous les dossiers nécessaires
    fs.writeFileSync(outputPath, pdfBuffer);
    console.log(`✅ PDF généré avec succès: ${outputPath}`);

    return pdfBuffer; // Pour Express: res.send(pdfBuffer)
  } catch (error) {
    console.error("❌ Erreur génération PDF:", error);
    throw error;
  }
}

// Export pour utilisation en module
module.exports = { PDFWasmGenerator, generatePDFFromTemplate };

// Si exécuté directement
if (require.main === module) {
  mainProcess().catch(console.error);
}
