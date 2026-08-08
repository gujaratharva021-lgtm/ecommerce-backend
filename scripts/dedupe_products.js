const BASE = "http://192.168.1.12:8081/api/v1";

async function main() {
  const idx = process.argv.indexOf("--token");
  const token = idx >= 0 ? process.argv[idx + 1] : null;
  if (!token) {
    console.error("Usage: node dedupe_products.js --token YOUR_ADMIN_TOKEN");
    process.exit(1);
  }
  const headers = { Authorization: `Bearer ${token}`, "Content-Type": "application/json" };

  // Fetch ALL products (paginated)
  let all = [];
  let page = 1;
  while (true) {
    const res = await fetch(`${BASE}/products/?limit=100&page=${page}`);
    const data = await res.json();
    all = all.concat(data.products || []);
    if (page >= (data.total_pages || 1)) break;
    page++;
  }
  console.log(`Total products fetched: ${all.length}`);

  // Group by category_id + normalized name
  const groups = {};
  for (const p of all) {
    const key = `${p.category_id}::${p.name.trim().toLowerCase()}`;
    if (!groups[key]) groups[key] = [];
    groups[key].push(p);
  }

  let toDelete = [];
  for (const key in groups) {
    const items = groups[key].sort((a, b) => b.id - a.id); // highest id first
    if (items.length > 1) {
      toDelete.push(...items.slice(1)); // keep first (highest id), delete rest
    }
  }

  console.log(`Duplicate groups found. Products to delete: ${toDelete.length}`);
  console.log(`Products to keep: ${all.length - toDelete.length}`);

  let ok = 0, fail = 0;
  for (const p of toDelete) {
    try {
      const res = await fetch(`${BASE}/admin/products/${p.id}`, { method: "DELETE", headers });
      if (res.ok) {
        ok++;
      } else {
        fail++;
        console.error(`  FAIL delete id ${p.id} (${p.name})`);
      }
    } catch (e) {
      fail++;
      console.error(`  FAIL delete id ${p.id}: ${e.message}`);
    }
  }

  console.log(`\nDone. Deleted: ${ok}, Failed: ${fail}`);
}

main().catch((e) => {
  console.error("Fatal error:", e);
  process.exit(1);
});
