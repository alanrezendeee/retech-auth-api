import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  ListToolsRequestSchema,
  CallToolRequestSchema
} from "@modelcontextprotocol/sdk/types.js";
import pg from "pg";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";

// ---------- ENV ----------
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

dotenv.config({
  path: path.join(__dirname, ".env")
});

if (!process.env.DATABASE_URL) {
  throw new Error("DATABASE_URL não definida");
}

// ---------- DB ----------
const { Pool } = pg;

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  statement_timeout: 5000,
  query_timeout: 5000,
  max: 5
});

// ---------- SECURITY ----------
function isForbiddenSql(sql) {
  const s = sql.trim().toLowerCase();

  return [
    "drop",
    "truncate",
    "alter",
    "grant",
    "revoke",
    "comment",
    "vacuum",
    "reindex"
  ].some(cmd => s.startsWith(cmd));
}

function isWriteSql(sql) {
  const s = sql.trim().toLowerCase();
  return (
    s.startsWith("insert") ||
    s.startsWith("update") ||
    s.startsWith("delete")
  );
}

// ---------- SERVER ----------
const server = new Server(
  { name: "postgres", version: "1.0.0" },
  { capabilities: { tools: {} } }
);

// ---------- TOOL ----------
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    {
      name: "execute_sql",
      description: "Executa SQL no PostgreSQL",
      inputSchema: {
        type: "object",
        properties: {
          sql: { type: "string" }
        },
        required: ["sql"]
      }
    }
  ]
}));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  if (name !== "execute_sql") {
    throw new Error("Tool inválida");
  }

  const sql = String(args?.sql || "").trim();
  if (!sql) throw new Error("SQL vazio");

  if (isForbiddenSql(sql)) {
    throw new Error("Comando não permitido");
  }

  if (isWriteSql(sql)) {
    console.log("⚠️ WRITE:", sql);
  }

  try {
    const result = await pool.query(sql);

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(result.rows)
        }
      ]
    };
  } catch (err) {
    return {
      content: [
        {
          type: "text",
          text: err.message
        }
      ],
      isError: true
    };
  }
});

// ---------- START ----------
const transport = new StdioServerTransport();
await server.connect(transport);

// ---------- SHUTDOWN ----------
process.on("SIGINT", async () => {
  await pool.end();
  process.exit(0);
});

process.on("SIGTERM", async () => {
  await pool.end();
  process.exit(0);
});