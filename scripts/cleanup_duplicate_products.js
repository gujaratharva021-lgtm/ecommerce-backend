/**
 * Removes duplicate products created by running bulk_import_products.js more than once.
 * Keeps the OLDEST copy (lowest id) of each (name + category_id) pair, deletes the rest.
 *
 * USAGE (from ecommerce-backend/scripts folder):
 *   node cleanup_duplicate_products.js --token YOUR_ADMIN_TOKEN --base http://192.168.1.12:8081/api/v1
 *
 * --base is optional, defaults to http://192.168.1.12:8081/api/v1
 */

function parseArgs() {
  const args = process.argv.slice(2);
  const out = { base: 'http://192.168.1.12:8081/api/v1' };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--token') out.token = args[++i];
    if (args[i] === '--base') out.base = args[++i];
  }
  return out;
}

async function main() {
  const { token, base } = parseArgs();
  if (!token) {
    console.error('Usage: node cleanup_duplicate_products.js --token YOUR_ADMIN_TOKEN [--base http://host:port/api/v1]');
    process.exit(1);
  }

  const headers = { Authorization: `Bearer ${token}` };

  console.log('Fetching all products (this may take a moment for a large catalog)...');
  let all = [];
  let page = 1;
  const limit = 200;
  while (true) {
    const res = await fetch(`${base}/products?page=${page}&limit=${limit}`);
    const data = await res.json();
    const batch = data.products || [];
    if (batch.length === 0) break;
    all = all.concat(batch);
    console.log(`  fetched page ${page}, total so far: ${all.length}`);
    if (batch.length < limit) break;
    page++;
  }

  console.log(`\nTotal products fetched: ${all.length}`);

  // Group by name + category_id
  const groups = new Map();
  for (const p of all) {
    const key = `${p.name}::${p.category_id}`;
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key).push(p);
  }

  const toDelete = [];
  for (const [key, items] of groups.entries()) {
    if (items.length > 1) {
      items.sort((a, b) => a.id - b.id);
      const [keep, ...dupes] = items;
      console.log(`  "${key}" -> keeping id ${keep.id}, deleting ${dupes.map((d) => d.id).join(', ')}`);
      toDelete.push(...dupes);
    }
  }

  console.log(`\nFound ${toDelete.length} duplicate products to delete.\n`);
  if (toDelete.length === 0) {
    console.log('Nothing to clean up. Done.');
    return;
  }

  let ok = 0, fail = 0;
  for (const p of toDelete) {
    try {
      const res = await fetch(`${base}/admin/products/${p.id}`, {
        method: 'DELETE',
        headers,
      });
      if (res.ok) {
        ok++;
        console.log(`  DELETED id ${p.id} (${p.name})`);
      } else {
        fail++;
        const data = await res.json().catch(() => ({}));
        console.error(`  FAIL id ${p.id}:`, data.error || res.status);
      }
    } catch (e) {
      fail++;
      console.error(`  FAIL id ${p.id}:`, e.message);
    }
  }

  console.log(`\nDone. Deleted: ${ok}, Failed: ${fail}`);
}

main().catch((e) => {
  console.error('Fatal error:', e);
  process.exit(1);
});
