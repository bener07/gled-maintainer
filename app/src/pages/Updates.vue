<template>
  <div>
    <!-- Stats row -->
    <div class="row mb-3">
      <div class="col-md-6">
        <card>
          <template slot="header">
            <h4 class="card-title">Nova Atualização</h4>
          </template>
          <form @submit.prevent="createUpdate">
            <base-input
              v-model="form.version"
              label="Versão"
              placeholder="ex: 1.2.3"
              required
            />
            <div class="form-group">
              <label>Changelog</label>
              <textarea
                v-model="form.changelog"
                class="form-control"
                rows="4"
                placeholder="Descreva as alterações..."
                required
              ></textarea>
            </div>
            <div class="form-group">
              <label>Agendar para (opcional)</label>
              <input
                type="datetime-local"
                v-model="form.scheduled_at"
                class="form-control"
              />
            </div>
            <button
              type="submit"
              class="btn btn-primary w-100"
              :disabled="creating"
            >
              {{ creating ? "A criar..." : "Criar Atualização" }}
            </button>
          </form>
        </card>
      </div>

      <div class="col-md-6">
        <card>
          <template slot="header">
            <div class="d-flex justify-content-between align-items-center">
              <h4 class="card-title">Atualizações</h4>
              <button class="btn btn-sm btn-primary" @click="refresh">
                <i class="tim-icons icon-refresh-02"></i>
              </button>
            </div>
          </template>

          <div v-if="loading" class="text-center py-4">A carregar...</div>
          <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
          <div v-else class="table-responsive">
            <table class="table tablesorter">
              <thead class="text-primary">
                <tr>
                  <th>Versão</th>
                  <th>Estado</th>
                  <th>Criada em</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="updates.length === 0">
                  <td colspan="4" class="text-center text-muted">
                    Nenhuma atualização criada.
                  </td>
                </tr>
                <tr v-for="u in updates" :key="u.id" @click="selectUpdate(u)" style="cursor:pointer">
                  <td>
                    <strong>{{ u.version }}</strong>
                    <div class="text-muted small">{{ truncate(u.changelog, 60) }}</div>
                  </td>
                  <td>
                    <span
                      class="badge"
                      :class="{
                        'badge-warning': u.status === 'pending',
                        'badge-success': u.status === 'applied',
                        'badge-danger': u.status === 'failed',
                      }"
                    >{{ u.status }}</span>
                  </td>
                  <td>{{ formatDate(u.created_at) }}</td>
                  <td>
                    <button
                      class="btn btn-sm btn-danger"
                      @click.stop="confirmDelete(u)"
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

    <!-- Detail panel -->
    <div class="row" v-if="selected">
      <div class="col-12">
        <card>
          <template slot="header">
            <div class="d-flex justify-content-between align-items-center">
              <h4 class="card-title">Detalhes — v{{ selected.version }}</h4>
              <button class="btn btn-sm btn-secondary" @click="selected = null">
                Fechar
              </button>
            </div>
          </template>
          <p><strong>Estado:</strong> {{ selected.status }}</p>
          <p v-if="selected.scheduled_at">
            <strong>Agendado para:</strong> {{ formatDate(selected.scheduled_at) }}
          </p>
          <p><strong>Changelog:</strong></p>
          <pre class="text-white" style="white-space: pre-wrap;">{{ selected.changelog }}</pre>
        </card>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <modal v-if="deleteTarget" @close="deleteTarget = null">
      <template slot="header">
        <h5 class="modal-title">Confirmar Remoção</h5>
      </template>
      <div>
        Remover atualização <strong>v{{ deleteTarget.version }}</strong>?
        Esta ação não pode ser desfeita.
      </div>
      <template slot="footer">
        <button class="btn btn-secondary" @click="deleteTarget = null">Cancelar</button>
        <button class="btn btn-danger" @click="deleteUpdate">Remover</button>
      </template>
    </modal>
  </div>
</template>

<script>
import { updatesAPI } from "@/services/api";
import Modal from "@/components/Modal.vue";

export default {
  name: "Updates",
  components: { Modal },
  data() {
    return {
      updates: [],
      loading: false,
      error: null,
      creating: false,
      selected: null,
      deleteTarget: null,
      form: {
        version: "",
        changelog: "",
        scheduled_at: "",
      },
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
        const res = await updatesAPI.list();
        this.updates = res.data.data || [];
      } catch (e) {
        this.error = e.response?.data?.message || "Erro ao carregar atualizações.";
      } finally {
        this.loading = false;
      }
    },
    async createUpdate() {
      this.creating = true;
      try {
        const payload = {
          version: this.form.version,
          changelog: this.form.changelog,
        };
        if (this.form.scheduled_at) {
          payload.scheduled_at = new Date(this.form.scheduled_at).toISOString();
        }
        await updatesAPI.create(payload);
        this.$notify({ message: "Atualização criada com sucesso.", type: "success" });
        this.form = { version: "", changelog: "", scheduled_at: "" };
        await this.refresh();
      } catch (e) {
        this.$notify({
          message: e.response?.data?.message || "Erro ao criar atualização.",
          type: "danger",
        });
      } finally {
        this.creating = false;
      }
    },
    selectUpdate(u) {
      this.selected = this.selected?.id === u.id ? null : u;
    },
    confirmDelete(u) {
      this.deleteTarget = u;
    },
    async deleteUpdate() {
      try {
        await updatesAPI.delete(this.deleteTarget.id);
        this.updates = this.updates.filter((u) => u.id !== this.deleteTarget.id);
        if (this.selected?.id === this.deleteTarget.id) this.selected = null;
        this.deleteTarget = null;
        this.$notify({ message: "Atualização removida.", type: "success" });
      } catch (e) {
        this.$notify({
          message: e.response?.data?.message || "Erro ao remover.",
          type: "danger",
        });
      }
    },
    formatDate(val) {
      if (!val) return "—";
      return new Date(val).toLocaleString("pt-PT");
    },
    truncate(str, len) {
      return str && str.length > len ? str.slice(0, len) + "…" : str;
    },
  },
};
</script>
