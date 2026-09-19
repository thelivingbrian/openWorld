(() => {
  const status = document.getElementById('world-operation-status');
  async function request(path, body) {
    const tokenResponse = await fetch('/api/csrf');
    if (!tokenResponse.ok) throw new Error(await tokenResponse.text());
    const { token } = await tokenResponse.json();
    const response = await fetch(path, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': token },
      body: JSON.stringify(body),
    });
    if (!response.ok) throw new Error(await response.text());
    return response.status === 204 ? null : response.json();
  }
  document.getElementById('create-world-form').addEventListener('submit', async event => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = form.querySelector('button');
    button.disabled = true;
    status.textContent = 'Creating world…';
    try {
      const fields = new FormData(form);
      const world = await request('/api/design/worlds', { name: fields.get('name'), seed: fields.get('seed') });
      location.href = '/design/?world=' + encodeURIComponent(world.id);
    } catch (error) { status.textContent = error.message; button.disabled = false; }
  });
  document.querySelectorAll('[data-action]').forEach(button => button.addEventListener('click', async () => {
    const { world, action } = button.dataset;
    if (action === 'stop' && !confirm('Shut down this world and disconnect its players? It will stay off until you launch it again.')) return;
    button.disabled = true;
    status.textContent = action === 'stop' ? 'Shutting down…' : 'Launching…';
    try { await request('/api/worlds/' + encodeURIComponent(world) + '/' + action, {}); location.reload(); }
    catch (error) { status.textContent = error.message; button.disabled = false; }
  }));
})();
