import { NextRequest, NextResponse } from "next/server";

export const runtime = "nodejs";

function upstreamBase(): string {
  return (process.env.API_UPSTREAM ?? "http://127.0.0.1:8080").replace(/\/$/, "");
}

const HOP_HEADERS = new Set([
  "connection",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailers",
  "transfer-encoding",
  "upgrade",
  "host",
]);

async function proxy(req: NextRequest, pathSegments: string[] | undefined) {
  const subpath = pathSegments?.join("/") ?? "";
  const url = new URL(req.url);
  const target = `${upstreamBase()}/api/${subpath}${url.search}`;

  const headers = new Headers();
  req.headers.forEach((value, key) => {
    if (HOP_HEADERS.has(key.toLowerCase())) return;
    headers.set(key, value);
  });

  let body: ArrayBuffer | undefined;
  if (!["GET", "HEAD"].includes(req.method)) {
    body = await req.arrayBuffer();
  }

  let res: Response;
  try {
    res = await fetch(target, {
      method: req.method,
      headers,
      body: body && body.byteLength > 0 ? body : undefined,
    });
  } catch (e) {
    const msg = e instanceof Error ? e.message : "fetch failed";
    return NextResponse.json(
      {
        error: "API proxy could not reach Go server",
        details: [
          `${msg}. Is the API running? Try API_UPSTREAM=${upstreamBase()} in frontend/.env.local`,
        ],
      },
      { status: 502 }
    );
  }

  const outHeaders = new Headers(res.headers);
  return new NextResponse(res.body, {
    status: res.status,
    statusText: res.statusText,
    headers: outHeaders,
  });
}

type Ctx = { params: Promise<{ path?: string[] }> };

export async function GET(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}

export async function POST(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}

export async function PUT(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}

export async function PATCH(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}

export async function DELETE(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}

export async function OPTIONS(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}
