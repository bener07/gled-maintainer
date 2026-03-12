<template>
  <div>
    <div class="row">
      <div class="col-12">
        <card>
          <template slot="header">
            <div class="d-flex justify-content-between align-items-center">
              <h4 class="card-title">Clientes</h4>
              <button class="btn btn-sm btn-primary" @click="refresh">
                <i class="tim-icons icon-refresh-02"></i> Atualizar
              </button>
            </div>
          </template>

          <div v-if="loading" class="text-center py-4">
            <span>A carregar...</span>
          </div>

          <div v-else-if="error" class="alert alert-danger">{{ error }}</div>

          <div v-else class="table-responsive">
            <table class="table tablesorter">
              <thead class="text-primary">
                <tr>
                  <th>ID</th>
                  <th>Nome</th>
                  <th>Hostname</th>
                  <th>OS</th>
                  <th>Versão</th>
                  <th>Estado</th>
                  <th>Último Heartbeat</th>
                  <th>Ações</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="clients.length === 0">
                  <td colspan="8" class="text-center text-muted">
                    Nenhum cliente registado.
                  </td>
                </tr>
                <tr v-for="client in clients" :key="client.id">
                  <td>{{ client.id }}</td>
                  <td>{{ client.name }}</td>
                  <td>{{ client.hostname }}</td>
                  <td>{{ client.os }}</td>
                  <td>{{ client.version }}</td>
                  <td>
                    <span
                      class="badge"
                      :class="{
                        'badge-success': client.status === 'online',
                        'badge-danger': client.status === 'offline',
                        'badge-secondary': client.status === 'unknown',
                      }"
                    >
                      {{ client.status }}
                    </span>
                  </td>
                  <td>{{ formatDate(client.last_heartbeat) }}</td>
                  <td>
                    <button
                      class="btn btn-sm btn-danger"
                      @click="confirmDelete(client)"
                      title="Remover"
                    >
                      <i class="tim-icons icon-simple-remove"></i>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </card>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <modal v-if="deleteTarget" @close="deleteTarget = null">
      <template slot="header">
        <h5 class="modal-title">Confirmar Remoção</h5>
      </template>
      <div>
        Tem a certeza que quer remover o cliente
        <strong>{{ deleteTarget.name }}</strong> ({{ deleteTarget.hostname }})?
      </div>
      <template slot="footer">
        <button class="btn btn-secondary" @click="deleteTarget = null">
          Cancelar
        </button>
        <button class="btn btn-danger" @click="deleteClient">Remover</button>
      </template>
    </modal>
  </div>
</template>

<script>
import { clientsAPI } from "@/services/api";
import Modal from "@/components/Modal.vue";

export default {
  name: "Clients",
  components: { Modal },
  data() {
    return {
      clients: [],
      loading: false,
      error: null,
      deleteTarget: null,
    };
  },
  mounted() {
    this.refresh();
  },
  methods: {
    async refresh() {
      this.loading = true;
      this.error = null;
      try {
        const res = await clientsAPI.list();
        this.clients = res.data.data || [];
      } catch (e) {
        this.error =
          e.response?.data?.message || "Erro ao carregar clientes.";
      } finally {
        this.loading = false;
      }
    },
    confirmDelete(client) {
      this.deleteTarget = client;
    },
    async deleteClient() {
      try {
        await clientsAPI.delete(this.deleteTarget.id);
        this.clients = this.clients.filter(
          (c) => c.id !== this.deleteTarget.id
        );
        this.deleteTarget = null;
        this.$notify({
          message: "Cliente removido com sucesso.",
          type: "success",
        });
      } catch (e) {
        this.$notify({
          message: e.response?.data?.message || "Erro ao remover cliente.",
          type: "danger",
        });
      }
    },
    formatDate(val) {
      if (!val) return "—";
      return new Date(val).toLocaleString("pt-PT");
    },
  },
};
</script>
