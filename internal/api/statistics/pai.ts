import { EChartsOption } from "echarts"
import { buildTooltip, filterData, renderChart, StatisticData } from "./statistics";

interface PaiData extends StatisticData {
	pai: number
}

export function InitPaiGraph(id: string, title: string, data: PaiData[]) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + id, () => {
		createPaiGraph(id, title, filterData(data))
	})
}

function createPaiGraph(id: string, title: string, data: PaiData[]) {
	const options: EChartsOption = {
		title: {
			text: title,
			textAlign: "center"
		},
		xAxis: [
			{
				type: "category",
				data: data.map(d => d.label),
			}
		],
		yAxis: [
			{
				type: "value",
			}
		],
		series: {
			type: "bar",
			data: data.map(d => d.pai),
			itemStyle: {
				color: (p) => {
					const val = (p.value ?? 0) as number

					if (val >= 100) {
						return "#7fd47f"
					}

					if (val >= 50) {
						return "#668a66"
					}

					return "#5a755a"
				}
			}
		},
		tooltip: {
			trigger: 'axis',
			axisPointer: {
				type: 'cross',
				label: {
					formatter: (p) => (
						p.value === null ? "" : (p.value as number).toLocaleString(undefined, { maximumFractionDigits: 0 })
					)
				}
			},
			formatter: (params) => buildTooltip(params as any, data, false)
		}
	}

	renderChart(id, options)
}