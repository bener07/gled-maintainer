<template>
  <div>
    <!-- Stats cards -->
    <div class="row">
      <div class="col-lg-3 col-md-6" v-for="stat in statCards" :key="stat.title">
        <card>
          <div class="d-flex justify-content-between align-items-center">
            <div>
              <p class="card-category">{{ stat.title }}</p>
              <h3 class="card-title">{{ stat.value }}</h3>
            </div>
            <div class="icon-big text-center" :class="'text-' + stat.type">
              <i :class="stat.icon"></i>
            </div>
          </div>
          <template slot="footer">
            <hr />
            <div class="stats"><i :class="stat.footerIcon"></i> {{ stat.footer }}</div>
          </template>
        </card>
      </div>
    </div>

    <!-- Clients and Updates summary -->
    <div class="row">
      <div class="col-lg-6 col-md-12">
        <card>
          <template slot="header">
            <div class="d-flex justify-content-between align-items-center">
              <h4 class="card-title">Clientes Recentes</h4>
              <router-link to="/clients" class="btn btn-sm btn-primary">
                Ver todos
              </router-link>
            </div>
          </template>
          <div v-if="loadingClients" class="text-center py-3">A carregar...</div>
          <div v-else class="table-responsive">
            <table class="table tablesorter">
              <thead class="text-primary">
                <tr>
                  <th>Nome</th>
                  <th>Hostname</th>
                  <th>Estado</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="recentClients.length === 0">
                  <td colspan="3" class="text-center text-muted">Nenhum cliente.</td>
                </tr>
                <tr v-for="c in recentClients" :key="c.id">
                  <td>{{ c.name }}</td>
                  <td>{{ c.hostname }}</td>
                  <td>
                    <span
                      class="badge"
                      :class="{
                        'badge-success': c.status === 'online',
                        'badge-danger': c.status === 'offline',
                        'badge-secondary': c.status === 'unknown',
                      }"
                    >{{ c.status }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </card>
      </div>

      <div class="col-lg-6 col-md-12">
        <card>
          <template slot="header">
            <div class="d-flex justify-content-between align-items-center">
              <h4 class="card-title">Últimas Atualizações</h4>
              <router-link to="/updates" class="btn btn-sm btn-primary">
                Gerir
              </router-link>
            </div>
          </template>
          <div v-if="loadingUpdates" class="text-center py-3">A carregar...</div>
          <div v-else class="table-responsive">
            <table class="table tablesorter">
              <thead class="text-primary">
                <tr>
                  <th>Versão</th>
                  <th>Estado</th>
                  <th>Data</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="recentUpdates.length === 0">
                  <td colspan="3" class="text-center text-muted">Nenhuma atualização.</td>
                </tr>
                <tr v-for="u in recentUpdates" :key="u.id">
                  <td>{{ u.version }}</td>
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
                </tr>
              </tbody>
            </table>
          </div>
        </card>
      </div>
    </div>
  </div>
</template>

<script>
import { statsAPI, clientsAPI, updatesAPI } from "@/services/api";

export default {
  components: {},
  data() {
    return {
      stats: {},
      recentClients: [],
      recentUpdates: [],
      loadingClients: false,
      loadingUpdates: false,
    };
  },
  computed: {
    statCards() {
      return [
        {
          title: "Total de Clientes",
          value: this.stats.total_clients || 0,
          type: "primary",
          icon: "tim-icons icon-laptop",
          footer: "Atualizado agora",
          footerIcon: "tim-icons icon-refresh-02",
        },
        {
          title: "Clientes Online",
          value: this.stats.online_clients || 0,
          type: "success",
          icon: "tim-icons icon-wifi",
          footer: "Em linha",
          footerIcon: "tim-icons icon-check-2",
        },
        {
          title: "Total de Atualizações",
          value: this.stats.total_updates || 0,
          type: "info",
          icon: "tim-icons icon-cloud-upload-94",
          footer: "Histórico",
          footerIcon: "tim-icons icon-calendar-60",
        },
        {
          title: "Pendentes",
          value: this.stats.pending_updates || 0,
          type: "warning",
          icon: "tim-icons icon-time-alarm",
          footer: "Por aplicar",
          footerIcon: "tim-icons icon-alert-circle-exc",
        },
      ];
    },
  },
  mounted() {
    this.loadAll();
  },
  methods: {
    async loadAll() {
      this.loadStats();
      this.loadClients();
      this.loadUpdates();
    },
    async loadStats() {
      try {
        const res = await statsAPI.get();
        this.stats = res.data.data || {};
      } catch {
        // non-critical
      }
    },
    async loadClients() {
      this.loadingClients = true;
      try {
        const res = await clientsAPI.list();
        this.recentClients = (res.data.data || []).slice(0, 5);
      } catch {
        // non-critical
      } finally {
        this.loadingClients = false;
      }
    },
    async loadUpdates() {
      this.loadingUpdates = true;
      try {
        const res = await updatesAPI.list();
        this.recentUpdates = (res.data.data || []).slice(0, 5);
      } catch {
        // non-critical
      } finally {
        this.loadingUpdates = false;
      }
    },
    formatDate(val) {
      if (!val) return "—";
      return new Date(val).toLocaleString("pt-PT");
    },
  },
};
</script>
<style></style>
