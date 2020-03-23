<template>
    <div>
        自动更新
        <i-switch v-model="autoUpdate"></i-switch>
        <div ref="mountNode"></div>
    </div>
</template>

<script>
let summaryES = null;
import G6 from "@antv/g6";
var graph = null;
export default {
    data() {
        return {
            autoUpdate: true,
            data: {}
        };
    },
    methods: {
        fetchSummary() {
            summaryES = new EventSource("/api/summary");
            summaryES.onmessage = evt => {
                if (!evt.data) return;
                let summary = JSON.parse(evt.data);
                summary.Address = location.hostname;
                if (!summary.Rooms) summary.Rooms = [];
                summary.Rooms.sort((a, b) =>
                    a.StreamPath > b.StreamPath ? 1 : -1
                );
                let d = this.addServer(summary);
                d.label = "🏠" + d.label;
                this.data = d;
            };
        },
        addServer(node) {
            let result = {
                id: node.Address,
                label: node.Address,
                description: `cpu:${node.CPUUsage >> 0}% mem:${node.Memory
                    .Usage >> 0}%`,
                shape: "modelRect",
                logoIcon: {
                    show: false
                },
                children: []
            };

            if (node.Rooms) {
                for (let i = 0; i < node.Rooms.length; i++) {
                    let room = node.Rooms[i];
                    let roomId = room.StreamPath;
                    let roomData = {
                        id: roomId,
                        label: room.StreamPath,
                        shape: "rect",
                        children: []
                    };
                    result.children.push(roomData);
                    if (room.SubscriberInfo) {
                        for (let j = 0; j < room.SubscriberInfo.length; j++) {
                            let subId = roomId + room.SubscriberInfo[j].ID;
                            roomData.children.push({
                                id: subId,
                                label: room.SubscriberInfo[j].ID
                            });
                        }
                    }
                }
            }
            if (node.Children) {
                for (let childId in node.Children) {
                    result.children.push(
                        this.addServer(node.Children[childId])
                    );
                }
            }
            return result;
        }
    },
    watch: {
        data(v) {
            if (graph && this.autoUpdate) {
                //graph.updateChild(v, "");
                graph.changeData(v); // 加载数据
                graph.fitView();
                //graph.read(v);
            }
        }
    },
    mounted() {
        this.fetchSummary();
        if (graph) return;
        graph = new G6.TreeGraph({
            linkCenter: true,
            // renderer: "svg",
            container: this.$refs.mountNode, // 指定挂载容器
            width: 800, // 图的宽度
            height: 500, // 图的高度
            modes: {
                default: [
                    "drag-canvas",
                    "zoom-canvas",
                    "click-select",
                    "drag-node"
                ]
            },
            animate: false,
            layout: {
                // type: "indeted",
                direction: "H"
            }
        });
        //graph.addChild(this.data, "");
        graph.read(this.data); // 加载数据
        graph.fitView();
    },
    deactivated() {
        summaryES.close();
    }
};
</script>

<style>
</style>