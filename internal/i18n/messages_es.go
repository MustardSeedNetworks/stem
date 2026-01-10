/*
 * The Stem - Spanish Messages (Mensajes en Espanol)
 *
 * All user-facing strings in Spanish.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package i18n

//nolint:gochecknoglobals // Static message catalog.
var spanishMessages = map[string]string{
	// Application
	"app.name":        "The Stem",
	"app.description": "Herramienta de Pruebas de Rendimiento de Red",
	"app.copyright":   "Copyright (c) 2025 Mustard Seed Networks. Todos los derechos reservados.",

	// Commands
	"cmd.reflect.name":    "reflect",
	"cmd.reflect.summary": "Ejecutar modo de reflejo de paquetes para pruebas remotas",
	"cmd.reflect.desc":    "Inicia el reflector de paquetes para devolver paquetes de prueba a su origen.",

	"cmd.test.name":    "test",
	"cmd.test.summary": "Ejecutar pruebas de rendimiento de red",
	"cmd.test.desc":    "Ejecuta pruebas RFC 2544, Y.1564, Y.1731 y otras pruebas de red.",

	"cmd.web.name":    "web",
	"cmd.web.summary": "Iniciar la interfaz web de Test Master",
	"cmd.web.desc":    "Lanza la interfaz web grafica para configuracion y monitoreo de pruebas.",

	"cmd.license.name":    "license",
	"cmd.license.summary": "Gestionar activacion de licencia",
	"cmd.license.desc":    "Activar, desactivar o verificar el estado de la licencia.",

	"cmd.version.name":    "version",
	"cmd.version.summary": "Mostrar informacion de version",
	"cmd.version.desc":    "Muestra version, informacion de compilacion y estado de licencia.",

	"cmd.help.name":    "help",
	"cmd.help.summary": "Obtener ayuda sobre comandos, pruebas y conceptos",
	"cmd.help.desc":    "Muestra documentacion detallada para cualquier tema.",

	"cmd.tutorial.name":    "tutorial",
	"cmd.tutorial.summary": "Tutoriales interactivos para aprender",
	"cmd.tutorial.desc":    "Guias paso a paso para tareas comunes.",

	"cmd.glossary.name":    "glossary",
	"cmd.glossary.summary": "Definiciones de terminologia de redes",
	"cmd.glossary.desc":    "Buscar terminos y conceptos de pruebas de red.",

	// Flags
	"flag.interface":   "Interfaz de red",
	"flag.interface.d": "Interfaz de red para transmision de paquetes (ej: eth0)",
	"flag.port":        "Numero de puerto",
	"flag.port.d":      "Numero de puerto TCP/UDP",
	"flag.verbose":     "Salida detallada",
	"flag.verbose.d":   "Habilitar registro detallado",
	"flag.duration":    "Duracion de prueba",
	"flag.duration.d":  "Cuanto tiempo ejecutar la prueba (en segundos)",
	"flag.output":      "Archivo de salida",
	"flag.output.d":    "Archivo para guardar resultados",
	"flag.config":      "Archivo de configuracion",
	"flag.config.d":    "Ruta al archivo de configuracion",

	// Test Categories
	"cat.rfc2544":      "RFC 2544",
	"cat.rfc2544.name": "Metodologia de Pruebas para Dispositivos de Interconexion de Red",
	"cat.rfc2544.desc": "Pruebas estandar para medir el rendimiento de dispositivos de red.",

	"cat.y1564":      "Y.1564",
	"cat.y1564.name": "Prueba de Activacion de Servicio Ethernet",
	"cat.y1564.desc": "Pruebas de activacion de servicio para ethernet de operador.",

	"cat.y1731":      "Y.1731",
	"cat.y1731.name": "OAM de Ethernet",
	"cat.y1731.desc": "Operaciones, Administracion y Mantenimiento para servicios ethernet.",

	"cat.rfc2889":      "RFC 2889",
	"cat.rfc2889.name": "Metodologia de Pruebas para Dispositivos de Conmutacion LAN",
	"cat.rfc2889.desc": "Pruebas especificas para switches de red.",

	"cat.rfc6349":      "RFC 6349",
	"cat.rfc6349.name": "Marco para Pruebas de Rendimiento TCP",
	"cat.rfc6349.desc": "Pruebas de rendimiento TCP considerando el comportamiento del protocolo.",

	"cat.mef":      "MEF",
	"cat.mef.name": "Metro Ethernet Forum",
	"cat.mef.desc": "Pruebas de certificacion de servicios ethernet de operador.",

	"cat.tsn":      "TSN",
	"cat.tsn.name": "Redes Sensibles al Tiempo",
	"cat.tsn.desc": "Redes deterministicas para aplicaciones industriales.",

	// Test Names
	"test.throughput":        "Prueba de Rendimiento",
	"test.throughput.desc":   "Encuentra la velocidad maxima sin perdida de paquetes",
	"test.latency":           "Prueba de Latencia",
	"test.latency.desc":      "Mide el retardo de paquetes a traves de la red",
	"test.frame_loss":        "Prueba de Perdida de Tramas",
	"test.frame_loss.desc":   "Mide la perdida de paquetes a varias tasas",
	"test.back_to_back":      "Prueba Consecutiva",
	"test.back_to_back.desc": "Mide la capacidad de manejo de rafagas",

	"test.y1564_config":      "Prueba de Configuracion Y.1564",
	"test.y1564_config.desc": "Valida el servicio en porcentajes de CIR",
	"test.y1564_perf":        "Prueba de Rendimiento Y.1564",
	"test.y1564_perf.desc":   "Verificacion extendida de rendimiento",

	"test.frame_delay":         "Prueba de Retardo de Trama",
	"test.frame_delay.desc":    "Medicion de retardo basada en OAM",
	"test.synthetic_loss":      "Medicion de Perdida Sintetica",
	"test.synthetic_loss.desc": "Medicion de perdida basada en OAM",

	// Status Messages
	"status.starting":   "Iniciando...",
	"status.running":    "Ejecutando",
	"status.completed":  "Completado",
	"status.failed":     "Fallido",
	"status.cancelled":  "Cancelado",
	"status.waiting":    "Esperando",
	"status.connecting": "Conectando...",

	// Results
	"result.pass":    "APROBADO",
	"result.fail":    "FALLIDO",
	"result.warning": "ADVERTENCIA",
	"result.info":    "INFO",

	// Units
	"unit.bps":     "bps",
	"unit.kbps":    "Kbps",
	"unit.mbps":    "Mbps",
	"unit.gbps":    "Gbps",
	"unit.pps":     "pps",
	"unit.kpps":    "Kpps",
	"unit.mpps":    "Mpps",
	"unit.ms":      "ms",
	"unit.us":      "us",
	"unit.ns":      "ns",
	"unit.percent": "%",
	"unit.bytes":   "bytes",

	// Errors
	"err.interface_required":  "Se requiere interfaz de red",
	"err.interface_not_found": "Interfaz no encontrada: %s",
	"err.test_type_required":  "Se requiere tipo de prueba",
	"err.test_type_invalid":   "Tipo de prueba invalido: %s",
	"err.license_required":    "Se requiere licencia valida para esta funcion",
	"err.license_expired":     "La licencia ha expirado",
	"err.connection_failed":   "Fallo la conexion al reflector",
	"err.permission_denied":   "Permiso denegado (intente ejecutar como root)",
	"err.port_in_use":         "El puerto %d ya esta en uso",
	"err.config_not_found":    "Archivo de configuracion no encontrado: %s",
	"err.config_invalid":      "Configuracion invalida: %s",

	// Prompts
	"prompt.continue":         "Presione Enter para continuar...",
	"prompt.confirm":          "Esta seguro? (s/n): ",
	"prompt.select_test":      "Seleccione un tipo de prueba:",
	"prompt.select_interface": "Seleccione una interfaz de red:",

	// Help
	"help.usage":     "Uso",
	"help.examples":  "Ejemplos",
	"help.flags":     "Opciones",
	"help.see_also":  "Ver Tambien",
	"help.technical": "Descripcion Tecnica",
	"help.simple":    "Explicacion Simple",
	"help.when_use":  "Cuando Usar",
	"help.when_not":  "Cuando No Usar",
	"help.tips":      "Consejos",
	"help.issues":    "Problemas Comunes",
	"help.params":    "Parametros",
	"help.metrics":   "Metricas de Salida",

	// UI Labels
	"ui.dashboard":  "Panel de Control",
	"ui.tests":      "Pruebas",
	"ui.results":    "Resultados",
	"ui.reflector":  "Reflector",
	"ui.settings":   "Configuracion",
	"ui.help":       "Ayuda",
	"ui.start":      "Iniciar",
	"ui.stop":       "Detener",
	"ui.cancel":     "Cancelar",
	"ui.save":       "Guardar",
	"ui.export":     "Exportar",
	"ui.refresh":    "Actualizar",
	"ui.filter":     "Filtrar",
	"ui.search":     "Buscar",
	"ui.loading":    "Cargando...",
	"ui.no_results": "No se encontraron resultados",
	"ui.error":      "Error",
	"ui.success":    "Exito",
	"ui.warning":    "Advertencia",

	// Time
	"time.now":       "Ahora",
	"time.today":     "Hoy",
	"time.yesterday": "Ayer",
	"time.last_week": "Semana Pasada",
	"time.ago":       "hace %s",
	"time.seconds":   "segundos",
	"time.minutes":   "minutos",
	"time.hours":     "horas",
	"time.days":      "dias",

	// Reflector
	"reflector.mode":    "Modo Reflector",
	"reflector.started": "Reflector iniciado en %s",
	"reflector.stopped": "Reflector detenido",
	"reflector.stats":   "Estadisticas del Reflector",
	"reflector.packets": "Paquetes Reflejados",
	"reflector.bytes":   "Bytes Reflejados",
	"reflector.rate":    "Tasa Actual",
	"reflector.uptime":  "Tiempo Activo",

	// License
	"license.status":      "Estado de Licencia",
	"license.tier":        "Nivel de Licencia",
	"license.valid_until": "Valido Hasta",
	"license.features":    "Funciones Disponibles",
	"license.activate":    "Activar Licencia",
	"license.deactivate":  "Desactivar Licencia",
	"license.trial":       "Modo de Prueba",
	"license.expired":     "Expirado",

	// Modules
	"module.reflector":   "Reflector",
	"module.benchmark":   "Benchmark",
	"module.servicetest": "Prueba de Servicio",
	"module.trafficgen":  "Generador de Trafico",
	"module.measure":     "Medir",
	"module.certify":     "Certificar",

	// Test Parameters - RFC 2544
	"param.frameSizes":      "Tamanos de Trama",
	"param.frameSizes.d":    "Tamanos de trama Ethernet en bytes a probar",
	"param.duration":        "Duracion",
	"param.duration.d":      "Duracion de la prueba en segundos",
	"param.resolution":      "Resolucion",
	"param.resolution.d":    "Porcentaje de resolucion de busqueda binaria",
	"param.maxLoss":         "Perdida Maxima",
	"param.maxLoss.d":       "Porcentaje maximo aceptable de perdida de tramas",
	"param.warmup":          "Calentamiento",
	"param.warmup.d":        "Periodo de calentamiento en segundos antes de las mediciones",
	"param.trials":          "Intentos",
	"param.trials.d":        "Numero de iteraciones de prueba por punto",
	"param.stepSize":        "Tamano de Paso",
	"param.stepSize.d":      "Tamano de paso de tasa para prueba de perdida de tramas",
	"param.bidirectional":   "Bidireccional",
	"param.bidirectional.d": "Ejecutar pruebas en ambas direcciones simultaneamente",

	// Test Parameters - Y.1564
	"param.cir":            "CIR",
	"param.cir.d":          "Tasa de Informacion Comprometida en Mbps",
	"param.eir":            "EIR",
	"param.eir.d":          "Tasa de Informacion Excedente en Mbps",
	"param.cbs":            "CBS",
	"param.cbs.d":          "Tamano de Rafaga Comprometida en KB",
	"param.ebs":            "EBS",
	"param.ebs.d":          "Tamano de Rafaga Excedente en KB",
	"param.vlanId":         "ID de VLAN",
	"param.vlanId.d":       "Identificador de VLAN para trafico etiquetado (0-4095)",
	"param.pcp":            "PCP",
	"param.pcp.d":          "Punto de Codigo de Prioridad para CoS 802.1p (0-7)",
	"param.colorAware":     "Consciente de Color",
	"param.colorAware.d":   "Habilitar condicionamiento de trafico consciente de color",
	"param.flrThreshold":   "Umbral FLR",
	"param.flrThreshold.d": "Umbral de aceptacion de Ratio de Perdida de Tramas",
	"param.fdThreshold":    "Umbral FD",
	"param.fdThreshold.d":  "Umbral de aceptacion de Retardo de Trama en ms",
	"param.fdvThreshold":   "Umbral FDV",
	"param.fdvThreshold.d": "Umbral de Variacion de Retardo de Trama en ms",

	// Test Parameters - TSN
	"param.maxLatencyNs":     "Latencia Maxima",
	"param.maxLatencyNs.d":   "Latencia maxima aceptable en nanosegundos",
	"param.maxJitterNs":      "Jitter Maximo",
	"param.maxJitterNs.d":    "Jitter maximo aceptable en nanosegundos",
	"param.requirePTPSync":   "Requerir Sincronizacion PTP",
	"param.requirePTPSync.d": "Requerir sincronizacion PTP antes de probar",
	"param.baseTimeNs":       "Tiempo Base",
	"param.baseTimeNs.d":     "Tiempo base para lista de control de puerta en nanosegundos",
	"param.cycleTimeNs":      "Tiempo de Ciclo",
	"param.cycleTimeNs.d":    "Tiempo de ciclo de puerta en nanosegundos",
	"param.trafficClass":     "Clase de Trafico",
	"param.trafficClass.d":   "Clase de trafico IEEE 802.1Q para trafico programado",

	// Test Parameters - TrafficGen
	"param.ratePct":           "Porcentaje de Tasa",
	"param.ratePct.d":         "Tasa de trafico como porcentaje de la tasa de linea",
	"param.streamId":          "ID de Flujo",
	"param.streamId.d":        "Identificador unico para flujo de trafico",
	"param.burstMode":         "Modo de Rafaga",
	"param.burstMode.d":       "Habilitar modo de trafico en rafagas",
	"param.burstSize":         "Tamano de Rafaga",
	"param.burstSize.d":       "Numero de tramas por rafaga",
	"param.interBurstGapUs":   "Intervalo entre Rafagas",
	"param.interBurstGapUs.d": "Intervalo entre rafagas en microsegundos",
	"param.srcMac":            "MAC de Origen",
	"param.srcMac.d":          "Direccion MAC de origen para tramas generadas",
	"param.dstMac":            "MAC de Destino",
	"param.dstMac.d":          "Direccion MAC de destino para tramas generadas",

	// TrafficGen Category
	"cat.trafficgen":      "TrafficGen",
	"cat.trafficgen.name": "Generacion de Trafico Personalizado",
	"cat.trafficgen.desc": "Generar patrones de trafico personalizados para pruebas especializadas",

	// Custom Stream Test
	"test.custom_stream":      "Flujo de Trafico Personalizado",
	"test.custom_stream.desc": "Generar patrones de trafico personalizados para pruebas especializadas",
}
