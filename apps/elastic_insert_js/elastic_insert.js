const { Client } = require('@elastic/elasticsearch');
const client = new Client({ node: 'http://localhost:9200' });

const indexName = 'index-2026';
const batchSize = 1000; // Jumlah dokumen per batch
const totalDocuments = 1000000; // Total dokumen yang akan diperbarui

function getRandomPhoneNumber() {
  // Menghasilkan nomor HP acak
  const prefix = '+62';
  const number = Math.floor(1000000000 + Math.random() * 9000000000);
  return `${prefix}${number}`;
}

function getRandomBoolean() {
  // Menghasilkan nilai boolean acak
  return Math.random() >= 0.5;
}

function getRandomBiometric() {
  // Menghasilkan nilai biometric acak (LOW/HIGH)
  return Math.random() >= 0.5 ? 'HIGH' : 'LOW';
}

async function createIndexIfNotExists() {
  const indexExists = await client.indices.exists({ index: indexName });
  if (!indexExists.body) {
    await client.indices.create({
      index: indexName,
      body: {
        mappings: {
          properties: {
            phone_number: { type: 'text' },
            is_vpc: { type: 'boolean' },
            biometric: { type: 'keyword' },
            is_changed: { type: 'boolean' },
            value: { type: 'float' },
            timestamp: { type: 'date' },
            latitude: { type: 'float' },
            longitude: { type: 'float' },
            distance: { type: 'float' },
            status: { type: 'keyword' },
            name: { type: 'text' },
            age: { type: 'integer' },
            score: { type: 'float' },
            level: { type: 'integer' }
          }
        }
      }
    });
    console.log(`Index ${indexName} created.`);
  } else {
    console.log(`Index ${indexName} already exists.`);
  }
}

async function run() {
  await createIndexIfNotExists();

  let bulkOps = [];
  for (let i = 1; i <= totalDocuments; i++) {
    // Buat data dokumen
    const doc = {
      phone_number: getRandomPhoneNumber(),
      is_vpc: getRandomBoolean(),
      biometric: getRandomBiometric(),
      is_changed: getRandomBoolean(),
      value: Math.random() * 18,
      timestamp: new Date().toISOString(),
      latitude: -6.2 + Math.random() * 0.2,
      longitude: 106.8 + Math.random() * 0.2,
      distance: Math.random() * 100,
      status: getRandomBoolean() ? 'active' : 'inactive',
      name: `name_${i}`,
      age: Math.floor(18 + Math.random() * 42),
      score: Math.random() * 100,
      level: Math.floor(Math.random() * 10)
    };

    // Tambahkan operasi index ke array bulk
    bulkOps.push({ index: { _index: indexName, _id: i } });
    bulkOps.push(doc);

    // Jika mencapai batchSize, kirim batch ke Elasticsearch
    if (i % batchSize === 0) {
      const response = await client.bulk({ refresh: true, body: bulkOps });
      if (response.errors) {
        console.error(`Errors occurred: ${JSON.stringify(response.items)}`);
      } else {
        console.log(`Indexed ${i} documents`);
      }
      bulkOps = []; // Kosongkan array bulkOps untuk batch berikutnya
    }
  }

  // Kirim sisa dokumen yang tidak terproses dalam batch terakhir
  if (bulkOps.length > 0) {
    const response = await client.bulk({ refresh: true, body: bulkOps });
    if (response.errors) {
      console.error(`Errors occurred: ${JSON.stringify(response.items)}`);
    } else {
      console.log(`Indexed remaining ${totalDocuments % batchSize} documents`);
    }
  }

  console.log('Finished indexing documents');
}

run().catch(console.error);
