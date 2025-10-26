#!/usr/bin/env node

/**
 * Generate SQL initialization script from ExerciseDB JSON data files.
 * Converts equipment, bodyparts, muscles, and exercises JSON data into
 * PostgreSQL-compatible SQL INSERT statements.
 */

const fs = require("node:fs");
const path = require("node:path");

// Helper function to escape SQL strings
function escapeSql(str) {
  if (str === null || str === undefined) return "NULL";
  return `'${String(str).replace(/'/g, "''")}'`;
}

// Helper function to convert array to PostgreSQL array format
function arrayToSql(arr) {
  if (!arr || !Array.isArray(arr)) return "NULL";
  return (
    "'{" +
    arr.map((item) => escapeSql(item).replace(/^'|'$/g, "")).join(",") +
    "}'"
  );
}

function generateSql(dataPath) {
  const sqlStatements = [];

  // Start with database and connection commands
  sqlStatements.push("-- ExerciseDB Data Initialization Script");
  sqlStatements.push(
    "-- Generated from https://github.com/ExerciseDB/exercisedb-api\n",
  );

  sqlStatements.push("\\c db\n");

  // Create equipment table
  sqlStatements.push("-- Create equipment table");
  sqlStatements.push("DROP TABLE IF EXISTS equipment CASCADE;");
  sqlStatements.push("CREATE TABLE equipment (");
  sqlStatements.push("  id SERIAL PRIMARY KEY,");
  sqlStatements.push("  name TEXT NOT NULL UNIQUE");
  sqlStatements.push(");\n");

  // Load equipment data
  const equipment = JSON.parse(
    fs.readFileSync(path.join(dataPath, "equipments.json"), "utf8"),
  );
  // Filter out "body weight" equipment
  const filteredEquipment = equipment.filter((e) => e.name !== "body weight");
  sqlStatements.push("-- Insert equipment data");
  sqlStatements.push("INSERT INTO equipment (name) VALUES");
  const equipmentValues = filteredEquipment.map(
    (e) => `  (${escapeSql(e.name)})`,
  );
  sqlStatements.push(`${equipmentValues.join(",\n")};\n`);

  // Create bodyparts table
  sqlStatements.push("-- Create bodyparts table");
  sqlStatements.push("DROP TABLE IF EXISTS bodyparts CASCADE;");
  sqlStatements.push("CREATE TABLE bodyparts (");
  sqlStatements.push("  id SERIAL PRIMARY KEY,");
  sqlStatements.push("  name TEXT NOT NULL UNIQUE");
  sqlStatements.push(");\n");

  // Load bodyparts data
  const bodyparts = JSON.parse(
    fs.readFileSync(path.join(dataPath, "bodyparts.json"), "utf8"),
  );
  sqlStatements.push("-- Insert bodyparts data");
  sqlStatements.push("INSERT INTO bodyparts (name) VALUES");
  const bodypartValues = bodyparts.map((b) => `  (${escapeSql(b.name)})`);
  sqlStatements.push(`${bodypartValues.join(",\n")};\n`);

  // Create muscles table
  sqlStatements.push("-- Create muscles table");
  sqlStatements.push("DROP TABLE IF EXISTS muscles CASCADE;");
  sqlStatements.push("CREATE TABLE muscles (");
  sqlStatements.push("  id SERIAL PRIMARY KEY,");
  sqlStatements.push("  name TEXT NOT NULL UNIQUE");
  sqlStatements.push(");\n");

  // Load muscles data
  const muscles = JSON.parse(
    fs.readFileSync(path.join(dataPath, "muscles.json"), "utf8"),
  );
  sqlStatements.push("-- Insert muscles data");
  sqlStatements.push("INSERT INTO muscles (name) VALUES");
  const muscleValues = muscles.map((m) => `  (${escapeSql(m.name)})`);
  sqlStatements.push(`${muscleValues.join(",\n")};\n`);

  // Create exercises table
  sqlStatements.push("-- Create exercises table");
  sqlStatements.push("DROP TABLE IF EXISTS exercises CASCADE;");
  sqlStatements.push("CREATE TABLE exercises (");
  sqlStatements.push("  id TEXT PRIMARY KEY,");
  sqlStatements.push("  name TEXT NOT NULL,");
  sqlStatements.push("  gif_url TEXT,");
  sqlStatements.push("  target_muscles TEXT[],");
  sqlStatements.push("  body_parts TEXT[],");
  sqlStatements.push("  equipment TEXT[],");
  sqlStatements.push("  secondary_muscles TEXT[],");
  sqlStatements.push("  instructions TEXT[]");
  sqlStatements.push(");\n");

  // Load exercises data
  const exercises = JSON.parse(
    fs.readFileSync(path.join(dataPath, "exercises.json"), "utf8"),
  );
  sqlStatements.push("-- Insert exercises data");

  // Process exercises in batches to avoid too long SQL statements
  const batchSize = 100;
  for (let i = 0; i < exercises.length; i += batchSize) {
    const batch = exercises.slice(i, i + batchSize);
    sqlStatements.push(
      "INSERT INTO exercises (id, name, gif_url, target_muscles, body_parts, equipment, secondary_muscles, instructions) VALUES",
    );

    const exerciseValues = batch.map((e) => {
      // Filter out "body weight" from equipment array
      const filteredEquipment = e.equipments
        ? e.equipments.filter((eq) => eq !== "body weight")
        : [];
      return `  (${escapeSql(e.exerciseId)}, ${escapeSql(e.name)}, ${escapeSql(e.gifUrl)}, ${arrayToSql(e.targetMuscles)}, ${arrayToSql(e.bodyParts)}, ${arrayToSql(filteredEquipment)}, ${arrayToSql(e.secondaryMuscles)}, ${arrayToSql(e.instructions)})`;
    });

    sqlStatements.push(`${exerciseValues.join(",\n")};\n`);
  }

  // Add summary statistics
  sqlStatements.push("-- Summary statistics");
  sqlStatements.push("SELECT");
  sqlStatements.push("  (SELECT COUNT(*) FROM equipment) as equipment_count,");
  sqlStatements.push("  (SELECT COUNT(*) FROM bodyparts) as bodypart_count,");
  sqlStatements.push("  (SELECT COUNT(*) FROM muscles) as muscle_count,");
  sqlStatements.push("  (SELECT COUNT(*) FROM exercises) as exercise_count;");

  return sqlStatements.join("\n");
}

// Main execution
const dataPath = process.argv[2] || "./src/data";

if (!fs.existsSync(dataPath)) {
  console.error(`Error: Data path '${dataPath}' does not exist`);
  process.exit(1);
}

try {
  const sql = generateSql(dataPath);
  console.log(sql);
} catch (error) {
  console.error("Error generating SQL:", error.message);
  process.exit(1);
}
