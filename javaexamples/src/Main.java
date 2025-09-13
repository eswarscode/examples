import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
      //  List<int[]> elms = merge(new int[][]{{1,3}, {2,6}, {8,10}, {15,18}});
        List<int[]> elms = merge(new int[][]{{1,4}, {4,5}});
        for (int[] elm : elms) {
            System.out.println(Arrays.toString(elm));
        }

    }

    public static List<int[]> merge(int[][] intervals) {
        if (intervals == null || intervals.length == 0) {
            return null;
        }
        Arrays.sort(intervals, (x, y) -> Integer.compare(x[0], y[0]));
        List<int[]> results = new ArrayList<>();

        int[] prev = intervals[0];
        for (int i = 1; i < intervals.length; i++) {
            int[] cur = intervals[i];
            if (cur[0] <= prev[1]) {
                prev[1] = Math.max(cur[1], prev[1]);
            } else {
                results.add(prev);
                prev = cur;
            }
        }
        results.add(prev);
         return results;
    }
}

