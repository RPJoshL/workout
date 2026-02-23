import { EChartsOption } from "echarts"
import { buildTooltip, filterData, renderChart, StatisticData } from "./statistics";

interface StepData extends StatisticData {
	steps: number
}

export function InitStepsGraph(id: string, title: string, data: StepData[]) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + id, () => {
		createStepsGraph(id, title, filterData(data))
	})
}

function createStepsGraph(id: string, title: string, data: StepData[]) {
	const options: EChartsOption = {
		title: {
			text: title,
			textAlign: "center"
		},
		xAxis: [
			{
				type: "category",
				data: data.map(d => d.label)
			}
		],
		yAxis: [
			{
				type: "value"
				
			}
		],
		series: {
			type: "bar",
			data: data.map(d => d.steps),
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