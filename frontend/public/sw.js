self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (event) =>
  event.waitUntil(self.clients.claim()),
);

function seen(id, remove = false) {
  return new Promise((resolve, reject) => {
    const opening = indexedDB.open("budget-push", 1);
    opening.onupgradeneeded = () => opening.result.createObjectStore("seen");
    opening.onerror = () => reject(opening.error);
    opening.onsuccess = () => {
      const db = opening.result;
      const tx = db.transaction("seen", "readwrite");
      const store = tx.objectStore("seen");
      let duplicate = false;
      if (remove) store.delete(id);
      else {
        const read = store.get(id);
        read.onsuccess = () => {
          duplicate = !!read.result;
          if (!duplicate) store.put(Date.now(), id);
        };
        const cursor = store.openCursor();
        cursor.onsuccess = () => {
          const row = cursor.result;
          if (row) {
            if (row.value < Date.now() - 7 * 86400000) row.delete();
            row.continue();
          }
        };
      }
      tx.oncomplete = () => {
        db.close();
        resolve(duplicate);
      };
      tx.onerror = () => {
        db.close();
        reject(tx.error);
      };
      tx.onabort = () => {
        db.close();
        reject(tx.error);
      };
    };
  });
}
self.addEventListener("push", (event) => {
  event.waitUntil(
    (async () => {
      if (!event.data) return;
      const data = event.data.json();
      if (!data.id || !data.title || (await seen(data.id))) return;
      try {
        await self.registration.showNotification(data.title, {
          body: data.body,
          tag: data.id,
          data: { url: "/" },
        });
      } catch (error) {
        await seen(data.id, true);
        throw error;
      }
    })(),
  );
});
self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  event.waitUntil(
    (async () => {
      const windows = await self.clients.matchAll({
        type: "window",
        includeUncontrolled: true,
      });
      for (const client of windows) {
        if (new URL(client.url).origin === self.location.origin) {
          await client.focus();
          return;
        }
      }
      await self.clients.openWindow("/");
    })(),
  );
});
